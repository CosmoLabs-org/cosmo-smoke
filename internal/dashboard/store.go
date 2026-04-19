package dashboard

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// RunRecord represents a stored smoke test run.
type RunRecord struct {
	ID               int64
	Project          string
	Timestamp        time.Time
	Total            int
	Passed           int
	Failed           int
	Skipped          int
	AllowedFailures  int
	DurationMs       int64
	Payload          string
}

// Store is the SQLite storage layer for dashboard data.
type Store struct {
	db               *sql.DB
	maxRunsPerProject int
}

// NewStore opens or creates a SQLite database at path.
// Use ":memory:" for an in-memory database.
func NewStore(path string, maxRunsPerProject int) (*Store, error) {
	if maxRunsPerProject <= 0 {
		maxRunsPerProject = 1000
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{db: db, maxRunsPerProject: maxRunsPerProject}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS runs (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			project          TEXT NOT NULL,
			timestamp        DATETIME DEFAULT CURRENT_TIMESTAMP,
			total            INTEGER,
			passed           INTEGER,
			failed           INTEGER,
			skipped          INTEGER,
			allowed_failures INTEGER DEFAULT 0,
			duration_ms      INTEGER,
			payload          TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_runs_project ON runs(project);
		CREATE INDEX IF NOT EXISTS idx_runs_timestamp ON runs(timestamp);
	`)
	return err
}

// InsertRun stores a smoke test result and prunes old runs.
func (s *Store) InsertRun(project string, payload []byte) (int64, error) {
	var data struct {
		Total           int `json:"total"`
		Passed          int `json:"passed"`
		Failed          int `json:"failed"`
		Skipped         int `json:"skipped"`
		AllowedFailures int `json:"allowed_failures"`
		DurationMs      int64 `json:"duration_ms"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return 0, fmt.Errorf("parse payload: %w", err)
	}

	res, err := s.db.Exec(`
		INSERT INTO runs (project, total, passed, failed, skipped, allowed_failures, duration_ms, payload)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		project, data.Total, data.Passed, data.Failed, data.Skipped,
		data.AllowedFailures, data.DurationMs, string(payload),
	)
	if err != nil {
		return 0, fmt.Errorf("insert run: %w", err)
	}

	id, _ := res.LastInsertId()
	s.pruneProject(project)
	return id, nil
}

// ProjectStatus is the latest status of a project.
type ProjectStatus struct {
	Name             string     `json:"name"`
	LatestStatus     string     `json:"latest_status"`
	TotalTests       int        `json:"total_tests"`
	Passed           int        `json:"passed"`
	Failed           int        `json:"failed"`
	LastRun          *time.Time `json:"last_run"`
	LastRunAgeSeconds int64     `json:"last_run_age_seconds"`
}

// GetProjects returns the latest status for all projects.
func (s *Store) GetProjects() ([]ProjectStatus, error) {
	rows, err := s.db.Query(`
		SELECT project, total, passed, failed, timestamp
		FROM runs r
		WHERE id = (
			SELECT id FROM runs r2
			WHERE r2.project = r.project
			ORDER BY timestamp DESC LIMIT 1
		)
		ORDER BY project
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []ProjectStatus
	now := time.Now()
	for rows.Next() {
		var p ProjectStatus
		var ts time.Time
		if err := rows.Scan(&p.Name, &p.TotalTests, &p.Passed, &p.Failed, &ts); err != nil {
			return nil, err
		}
		p.LastRun = &ts
		p.LastRunAgeSeconds = int64(now.Sub(ts).Seconds())
		if p.Failed > 0 {
			p.LatestStatus = "failing"
		} else {
			p.LatestStatus = "healthy"
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetProjectHistory returns recent runs for a project.
func (s *Store) GetProjectHistory(project string, limit int) ([]RunRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, project, timestamp, total, passed, failed, skipped, allowed_failures, duration_ms, payload
		FROM runs WHERE project = ?
		ORDER BY timestamp DESC LIMIT ?`,
		project, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []RunRecord
	for rows.Next() {
		var r RunRecord
		if err := rows.Scan(&r.ID, &r.Project, &r.Timestamp, &r.Total, &r.Passed,
			&r.Failed, &r.Skipped, &r.AllowedFailures, &r.DurationMs, &r.Payload); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) pruneProject(project string) {
	s.db.Exec(`
		DELETE FROM runs WHERE project = ? AND id NOT IN (
			SELECT id FROM runs WHERE project = ?
			ORDER BY timestamp DESC LIMIT ?
		)`,
		project, project, s.maxRunsPerProject,
	)
}
