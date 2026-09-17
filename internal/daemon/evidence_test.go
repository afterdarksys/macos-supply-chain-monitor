package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPackageIdentity(t *testing.T) {
	name, version := packageIdentity("/opt/homebrew/Cellar/tool/1.2.3/bin/tool", "homebrew")
	if name != "tool" || version != "1.2.3" {
		t.Fatalf("%s %s", name, version)
	}
	root := filepath.Join(t.TempDir(), "node_modules", "@scope", "tool")
	os.MkdirAll(root, 0700)
	os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"version":"2.0.1"}`), 0600)
	name, version = packageIdentity(filepath.Join(root, "index.js"), "npm")
	if name != "@scope/tool" || version != "2.0.1" {
		t.Fatalf("%s %s", name, version)
	}
}

func TestStaleHeartbeatRejected(t *testing.T) {
	db := filepath.Join(t.TempDir(), "events.db")
	data, _ := json.Marshal(Status{PID: 1, Updated: time.Now().Add(-time.Hour)})
	os.WriteFile(db+".status.json", data, 0600)
	if _, err := ReadStatus(db); err == nil {
		t.Fatal("stale process reported running")
	}
}
