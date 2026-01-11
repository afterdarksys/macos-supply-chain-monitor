package daemon

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/afterdark/supply-chain-monitor/internal/db"
	"github.com/afterdark/supply-chain-monitor/internal/watcher"
	"github.com/fsnotify/fsnotify"
)

type Daemon struct {
	store    *db.Store
	watchers []*watcher.Watcher
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func New(dbPath string) (*Daemon, error) {
	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	store, err := db.NewStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize store: %w", err)
	}

	return &Daemon{
		store:    store,
		stopChan: make(chan struct{}),
	}, nil
}

func (d *Daemon) Start() error {
	log.Println("Daemon starting...")

	// Initialize watchers for different paths
	watchPaths := []string{
		"/usr/local/Cellar",           // Homebrew Intel
		"/opt/homebrew/Cellar",        // Homebrew Apple Silicon
		"/usr/local/bin",
		"/opt/homebrew/bin",
	}

	// Check which paths exist
	for _, path := range watchPaths {
		if _, err := os.Stat(path); err == nil {
			w, err := d.startWatcher(path)
			if err != nil {
				log.Printf("Warning: failed to watch %s: %v", path, err)
				continue
			}
			d.watchers = append(d.watchers, w)
		}
	}

	if len(d.watchers) == 0 {
		return fmt.Errorf("no valid paths to watch")
	}

	log.Printf("Watching %d paths", len(d.watchers))
	return nil
}

func (d *Daemon) startWatcher(path string) (*watcher.Watcher, error) {
	w, err := watcher.New(path)
	if err != nil {
		return nil, err
	}

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.processEvents(w)
	}()

	return w, nil
}

func (d *Daemon) processEvents(w *watcher.Watcher) {
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			d.handleFileSystemEvent(event)

		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)

		case <-d.stopChan:
			return
		}
	}
}

func (d *Daemon) handleFileSystemEvent(event fsnotify.Event) {
	// Only care about new files and modifications
	if event.Op&fsnotify.Create == 0 && event.Op&fsnotify.Write == 0 {
		return
	}

	// Determine package manager from path
	pkgManager := d.detectPackageManager(event.Name)
	if pkgManager == "" {
		return
	}

	// Extract package name from path
	pkgName := d.extractPackageName(event.Name, pkgManager)

	// Calculate risk score
	riskScore := d.calculateRiskScore(event.Name, pkgManager)

	// Record event
	dbEvent := db.Event{
		EventType:      "install",
		PackageManager: pkgManager,
		PackageName:    pkgName,
		PackageVersion: "unknown", // TODO: Extract version
		RiskScore:      riskScore,
		Details: map[string]interface{}{
			"path":      event.Name,
			"operation": event.Op.String(),
		},
	}

	eventID, err := d.store.RecordEvent(dbEvent)
	if err != nil {
		log.Printf("Failed to record event: %v", err)
		return
	}

	if riskScore >= 70 {
		log.Printf("⚠️  HIGH RISK: %s installed %s (score: %d, event: %d)",
			pkgManager, pkgName, riskScore, eventID)
	} else {
		log.Printf("✓ %s installed %s (score: %d)", pkgManager, pkgName, riskScore)
	}
}

func (d *Daemon) detectPackageManager(path string) string {
	switch {
	case filepath.HasPrefix(path, "/usr/local/Cellar"),
		filepath.HasPrefix(path, "/opt/homebrew/Cellar"):
		return "homebrew"
	case filepath.HasPrefix(path, "/usr/local/lib/node_modules"),
		filepath.HasPrefix(path, "/opt/homebrew/lib/node_modules"):
		return "npm"
	default:
		return ""
	}
}

func (d *Daemon) extractPackageName(path, pkgManager string) string {
	switch pkgManager {
	case "homebrew":
		// /usr/local/Cellar/package-name/version/...
		// Extract package-name
		parts := filepath.SplitList(path)
		for i, part := range parts {
			if part == "Cellar" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
		// Fallback
		return filepath.Base(filepath.Dir(filepath.Dir(path)))
	case "npm":
		return filepath.Base(filepath.Dir(path))
	default:
		return filepath.Base(path)
	}
}

func (d *Daemon) calculateRiskScore(path, pkgManager string) int {
	score := 0

	// Check if file is executable
	info, err := os.Stat(path)
	if err == nil && info.Mode()&0111 != 0 {
		score += 10
	}

	// Check file size (very large binaries are suspicious)
	if err == nil && info.Size() > 50*1024*1024 { // > 50MB
		score += 20
	}

	// TODO: Check code signature
	// TODO: Check if creates LaunchAgent
	// TODO: Check if makes network connections
	// TODO: Entropy analysis

	return score
}

func (d *Daemon) Stop() error {
	log.Println("Daemon stopping...")

	close(d.stopChan)

	// Close all watchers
	for _, w := range d.watchers {
		w.Close()
	}

	// Wait for goroutines
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All watchers stopped")
	case <-time.After(5 * time.Second):
		log.Println("Timeout waiting for watchers to stop")
	}

	// Close database
	if err := d.store.Close(); err != nil {
		return fmt.Errorf("failed to close store: %w", err)
	}

	log.Println("Daemon stopped")
	return nil
}
