package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type Event struct {
	ID             int
	Timestamp      time.Time
	EventType      string
	PackageManager string
	PackageName    string
	PackageVersion string
	RiskScore      int
	Details        map[string]interface{}
}

type Binary struct {
	ID              int
	Path            string
	SHA256          string
	CodeSignID      string
	FirstSeen       time.Time
	LastModified    time.Time
	PackageSource   string
	RiskFlags       map[string]interface{}
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{db: db}
	if err := store.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		event_type TEXT NOT NULL,
		package_manager TEXT,
		package_name TEXT,
		package_version TEXT,
		risk_score INTEGER DEFAULT 0,
		details JSON
	);

	CREATE TABLE IF NOT EXISTS binaries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT UNIQUE NOT NULL,
		sha256 TEXT,
		code_sign_identity TEXT,
		first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_modified DATETIME DEFAULT CURRENT_TIMESTAMP,
		package_source TEXT,
		risk_flags JSON
	);

	CREATE TABLE IF NOT EXISTS persistence (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		path TEXT NOT NULL,
		target_binary TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_by_event INTEGER REFERENCES events(id),
		is_suspicious BOOLEAN DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_risk ON events(risk_score);
	CREATE INDEX IF NOT EXISTS idx_binaries_path ON binaries(path);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) RecordEvent(e Event) (int64, error) {
	detailsJSON, err := json.Marshal(e.Details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	result, err := s.db.Exec(`
		INSERT INTO events (event_type, package_manager, package_name, package_version, risk_score, details)
		VALUES (?, ?, ?, ?, ?, ?)
	`, e.EventType, e.PackageManager, e.PackageName, e.PackageVersion, e.RiskScore, string(detailsJSON))

	if err != nil {
		return 0, fmt.Errorf("failed to record event: %w", err)
	}

	return result.LastInsertId()
}

func (s *Store) GetRecentEvents(duration time.Duration, limit int, riskFilter string) ([]Event, error) {
	since := time.Now().Add(-duration)

	query := `
		SELECT id, timestamp, event_type, package_manager, package_name, package_version, risk_score, details
		FROM events
		WHERE timestamp >= ?
	`

	args := []interface{}{since}

	if riskFilter != "" {
		switch riskFilter {
		case "high":
			query += " AND risk_score >= 70"
		case "medium":
			query += " AND risk_score >= 40 AND risk_score < 70"
		case "low":
			query += " AND risk_score < 40"
		}
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var detailsJSON string

		err := rows.Scan(&e.ID, &e.Timestamp, &e.EventType, &e.PackageManager,
			&e.PackageName, &e.PackageVersion, &e.RiskScore, &detailsJSON)
		if err != nil {
			return nil, err
		}

		if detailsJSON != "" {
			json.Unmarshal([]byte(detailsJSON), &e.Details)
		}

		events = append(events, e)
	}

	return events, nil
}

func (s *Store) RecordBinary(b Binary) error {
	riskFlagsJSON, _ := json.Marshal(b.RiskFlags)

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO binaries (path, sha256, code_sign_identity, package_source, risk_flags, last_modified)
		VALUES (?, ?, ?, ?, ?, ?)
	`, b.Path, b.SHA256, b.CodeSignID, b.PackageSource, string(riskFlagsJSON), time.Now())

	return err
}

func (s *Store) GetBinary(path string) (*Binary, error) {
	var b Binary
	var riskFlagsJSON string

	err := s.db.QueryRow(`
		SELECT id, path, sha256, code_sign_identity, first_seen, last_modified, package_source, risk_flags
		FROM binaries WHERE path = ?
	`, path).Scan(&b.ID, &b.Path, &b.SHA256, &b.CodeSignID, &b.FirstSeen, &b.LastModified, &b.PackageSource, &riskFlagsJSON)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if riskFlagsJSON != "" {
		json.Unmarshal([]byte(riskFlagsJSON), &b.RiskFlags)
	}

	return &b, nil
}

func (s *Store) RecordPersistence(pType, path, targetBinary string, eventID int64, suspicious bool) error {
	_, err := s.db.Exec(`
		INSERT INTO persistence (type, path, target_binary, created_by_event, is_suspicious)
		VALUES (?, ?, ?, ?, ?)
	`, pType, path, targetBinary, eventID, suspicious)

	return err
}

func (s *Store) GetStats() (map[string]int, error) {
	stats := make(map[string]int)

	// Total events
	var totalEvents int
	err := s.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&totalEvents)
	if err != nil {
		return nil, err
	}
	stats["total_events"] = totalEvents

	// High risk events (last 7 days)
	weekAgo := time.Now().AddDate(0, 0, -7)
	var highRisk7d int
	err = s.db.QueryRow("SELECT COUNT(*) FROM events WHERE risk_score >= 70 AND timestamp >= ?", weekAgo).
		Scan(&highRisk7d)
	if err != nil {
		return nil, err
	}
	stats["high_risk_7d"] = highRisk7d

	// Total binaries tracked
	var totalBinaries int
	err = s.db.QueryRow("SELECT COUNT(*) FROM binaries").Scan(&totalBinaries)
	if err != nil {
		return nil, err
	}
	stats["total_binaries"] = totalBinaries

	// Persistence items
	var persistenceItems int
	err = s.db.QueryRow("SELECT COUNT(*) FROM persistence").Scan(&persistenceItems)
	if err != nil {
		return nil, err
	}
	stats["persistence_items"] = persistenceItems

	return stats, nil
}
