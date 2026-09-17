package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func packageIdentity(path, manager string) (string, string) {
	parts := strings.Split(filepath.ToSlash(filepath.Clean(path)), "/")
	for i, part := range parts {
		if manager == "homebrew" && part == "Cellar" && i+2 < len(parts) {
			return parts[i+1], parts[i+2]
		}
		if manager == "npm" && part == "node_modules" && i+1 < len(parts) {
			end := i + 2
			if strings.HasPrefix(parts[i+1], "@") && end < len(parts) {
				end++
			}
			name := strings.Join(parts[i+1:end], "/")
			manifest := strings.Join(parts[:end], "/") + "/package.json"
			if f, err := os.Open(manifest); err == nil {
				defer f.Close()
				var metadata struct {
					Version string `json:"version"`
				}
				if json.NewDecoder(io.LimitReader(f, 1<<20)).Decode(&metadata) == nil && metadata.Version != "" {
					return name, metadata.Version
				}
			}
			return name, "unknown"
		}
	}
	return filepath.Base(path), "unknown"
}

func fileRisk(path string) int {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return 0
	}
	score := 0
	if info.Mode()&0111 != 0 {
		score += 10
	}
	if info.Size() > 50<<20 {
		score += 20
	}
	if strings.Contains(filepath.ToSlash(path), "/LaunchAgents/") || strings.Contains(filepath.ToSlash(path), "/LaunchDaemons/") {
		score += 30
	}
	file, err := os.Open(path)
	if err != nil {
		return score
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 1<<20))
	if err != nil {
		return score
	}
	if len(data) > 4096 {
		var counts [256]int
		for _, b := range data {
			counts[b]++
		}
		entropy := 0.0
		for _, n := range counts {
			if n > 0 {
				p := float64(n) / float64(len(data))
				entropy -= p * math.Log2(p)
			}
		}
		if entropy > 7.5 {
			score += 20
		}
	}
	if runtime.GOOS == "darwin" && len(data) >= 4 && ((data[0] == 0xcf && data[1] == 0xfa) || (data[0] == 0xca && data[1] == 0xfe)) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if exec.CommandContext(ctx, "/usr/bin/codesign", "--verify", "--strict", path).Run() != nil {
			score += 25
		}
	}
	return min(score, 100)
}

type Status struct {
	PID      int       `json:"pid"`
	Updated  time.Time `json:"updated"`
	Watchers int       `json:"watchers"`
}

func ReadStatus(database string) (Status, error) {
	var status Status
	data, err := os.ReadFile(database + ".status.json")
	if err != nil {
		return status, err
	}
	if err = json.Unmarshal(data, &status); err != nil {
		return status, err
	}
	if status.PID <= 0 || time.Since(status.Updated) > 30*time.Second || time.Until(status.Updated) > 5*time.Second {
		return status, fmt.Errorf("daemon heartbeat is stale")
	}
	return status, nil
}

func (d *Daemon) heartbeat() {
	defer d.wg.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		data, _ := json.Marshal(Status{PID: os.Getpid(), Updated: time.Now().UTC(), Watchers: len(d.watchers)})
		if err := os.WriteFile(d.statusPath+".tmp", data, 0600); err == nil {
			_ = os.Rename(d.statusPath+".tmp", d.statusPath)
		}
		select {
		case <-d.stopChan:
			return
		case <-ticker.C:
		}
	}
}
