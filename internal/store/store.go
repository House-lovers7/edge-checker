package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/House-lovers7/edge-checker/internal/output"
)

const defaultDBPath = "results/history.db"

// Store manages SQLite storage for test results.
type Store struct {
	db *sql.DB
}

// Open opens or creates the SQLite database.
func Open(dbPath string) (*Store, error) {
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for database: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			scenario_name TEXT NOT NULL,
			description TEXT,
			started_at TEXT NOT NULL,
			ended_at TEXT NOT NULL,
			duration TEXT NOT NULL,
			interrupted INTEGER DEFAULT 0,
			mode TEXT NOT NULL,
			target_url TEXT NOT NULL,
			method TEXT NOT NULL,
			profile TEXT,
			environment TEXT,
			rps INTEGER,
			concurrency INTEGER,
			total_requests INTEGER,
			success_count INTEGER,
			error_count INTEGER,
			avg_latency_ms REAL,
			p50_latency_ms REAL,
			p95_latency_ms REAL,
			p99_latency_ms REAL,
			verdict_overall TEXT,
			result_json TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	return err
}

// Save stores a test result in the database.
func (s *Store) Save(result *output.Result) (int64, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal result: %w", err)
	}

	verdictOverall := ""
	if result.Verdict != nil {
		verdictOverall = string(result.Verdict.Overall)
	}

	interrupted := 0
	if result.Interrupted {
		interrupted = 1
	}

	res, err := s.db.Exec(`
		INSERT INTO runs (
			scenario_name, description, started_at, ended_at, duration, interrupted,
			mode, target_url, method, profile, environment, rps, concurrency,
			total_requests, success_count, error_count,
			avg_latency_ms, p50_latency_ms, p95_latency_ms, p99_latency_ms,
			verdict_overall, result_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		result.ScenarioName,
		result.Description,
		result.StartedAt.Format(time.RFC3339),
		result.EndedAt.Format(time.RFC3339),
		result.Duration,
		interrupted,
		result.Execution.Mode,
		result.Target.BaseURL+result.Target.Path,
		result.Target.Method,
		result.Target.Profile,
		result.Execution.Environment,
		result.Execution.RPS,
		result.Execution.Concurrency,
		result.Summary.TotalRequests,
		result.Summary.SuccessCount,
		result.Summary.ErrorCount,
		result.Summary.AvgLatencyMs,
		result.Summary.P50LatencyMs,
		result.Summary.P95LatencyMs,
		result.Summary.P99LatencyMs,
		verdictOverall,
		string(resultJSON),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to save result: %w", err)
	}

	return res.LastInsertId()
}

// RunSummary is a brief summary of a stored run for listing.
type RunSummary struct {
	ID            int64
	ScenarioName  string
	StartedAt     string
	Duration      string
	Mode          string
	TargetURL     string
	TotalRequests int
	Verdict       string
}

// List returns recent run summaries.
func (s *Store) List(limit int) ([]RunSummary, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT id, scenario_name, started_at, duration, mode, target_url,
		       total_requests, COALESCE(verdict_overall, '-')
		FROM runs
		ORDER BY id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query runs: %w", err)
	}
	defer rows.Close()

	var summaries []RunSummary
	for rows.Next() {
		var r RunSummary
		if err := rows.Scan(&r.ID, &r.ScenarioName, &r.StartedAt, &r.Duration,
			&r.Mode, &r.TargetURL, &r.TotalRequests, &r.Verdict); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		summaries = append(summaries, r)
	}

	return summaries, rows.Err()
}

// Get retrieves the full result JSON for a specific run.
func (s *Store) Get(id int64) (*output.Result, error) {
	var resultJSON string
	err := s.db.QueryRow(`SELECT result_json FROM runs WHERE id = ?`, id).Scan(&resultJSON)
	if err != nil {
		return nil, fmt.Errorf("run #%d not found: %w", id, err)
	}

	var result output.Result
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse stored result: %w", err)
	}

	return &result, nil
}
