package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nmashchenko/aegis-cli/internal/models"
	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func New(path string) (*DB, error) {
	if path != ":memory:" {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			name             TEXT NOT NULL,
			started_at       DATETIME NOT NULL,
			ended_at         DATETIME,
			duration_seconds INTEGER,
			limit_seconds    INTEGER
		);

		CREATE TABLE IF NOT EXISTS urges (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			task_id   INTEGER,
			FOREIGN KEY (task_id) REFERENCES tasks(id)
		);
	`)
	return err
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, ".aegis", "aegis.db"), nil
}

func (db *DB) CreateTask(name string, limitSeconds *int64) (int64, error) {
	result, err := db.conn.Exec(
		"INSERT INTO tasks (name, started_at, limit_seconds) VALUES (?, ?, ?)",
		name, time.Now().UTC(), limitSeconds,
	)
	if err != nil {
		return 0, fmt.Errorf("insert task: %w", err)
	}
	return result.LastInsertId()
}

// parseTime tries multiple time formats that SQLite may return.
func parseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %q", s)
}

func (db *DB) GetActiveTask() (*models.Task, error) {
	row := db.conn.QueryRow(
		"SELECT id, name, started_at, ended_at, duration_seconds, limit_seconds FROM tasks WHERE ended_at IS NULL LIMIT 1",
	)

	var task models.Task
	var startedAt string
	var endedAt sql.NullString
	var durationSeconds sql.NullInt64
	var limitSeconds sql.NullInt64

	err := row.Scan(&task.ID, &task.Name, &startedAt, &endedAt, &durationSeconds, &limitSeconds)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan active task: %w", err)
	}

	task.StartedAt, err = parseTime(startedAt)
	if err != nil {
		return nil, fmt.Errorf("parse started_at: %w", err)
	}

	if endedAt.Valid {
		t, err := parseTime(endedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse ended_at: %w", err)
		}
		task.EndedAt = &t
	}

	if durationSeconds.Valid {
		task.DurationSeconds = &durationSeconds.Int64
	}
	if limitSeconds.Valid {
		task.LimitSeconds = &limitSeconds.Int64
	}

	return &task, nil
}

func (db *DB) StopTask(id int64) error {
	// Read started_at and compute duration in Go, because SQLite's julianday()
	// cannot parse the timestamp format that modernc.org/sqlite produces.
	var startedAtStr string
	err := db.conn.QueryRow("SELECT started_at FROM tasks WHERE id = ? AND ended_at IS NULL", id).Scan(&startedAtStr)
	if err != nil {
		return fmt.Errorf("get task start time: %w", err)
	}

	startedAt, err := parseTime(startedAtStr)
	if err != nil {
		return fmt.Errorf("parse started_at: %w", err)
	}

	now := time.Now().UTC()
	durationSecs := int64(now.Sub(startedAt).Seconds())

	_, err = db.conn.Exec(
		`UPDATE tasks SET ended_at = ?, duration_seconds = ? WHERE id = ? AND ended_at IS NULL`,
		now, durationSecs, id,
	)
	if err != nil {
		return fmt.Errorf("stop task: %w", err)
	}
	return nil
}

func (db *DB) CreateUrge(taskID *int64) error {
	_, err := db.conn.Exec(
		"INSERT INTO urges (timestamp, task_id) VALUES (?, ?)",
		time.Now().UTC(), taskID,
	)
	if err != nil {
		return fmt.Errorf("insert urge: %w", err)
	}
	return nil
}

func (db *DB) GetUrgeCountForTask(taskID int64) (int, error) {
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM urges WHERE task_id = ?", taskID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count urges for task: %w", err)
	}
	return count, nil
}

func (db *DB) CountUrges(start, end time.Time) (int, error) {
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM urges WHERE timestamp >= ? AND timestamp < ?",
		start.UTC(), end.UTC(),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count urges: %w", err)
	}
	return count, nil
}

func (db *DB) GetStats(start, end time.Time) (*models.Stats, error) {
	stats := &models.Stats{}

	err := db.conn.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(duration_seconds), 0)
		 FROM tasks
		 WHERE ended_at IS NOT NULL
		   AND started_at >= ? AND started_at < ?`,
		start.UTC(), end.UTC(),
	).Scan(&stats.TasksCompleted, &stats.TotalFocusSecs)
	if err != nil {
		return nil, fmt.Errorf("query task stats: %w", err)
	}

	if stats.TasksCompleted > 0 {
		stats.AvgTaskSecs = stats.TotalFocusSecs / int64(stats.TasksCompleted)
	}

	stats.UrgesLogged, err = db.CountUrges(start, end)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetRecentTasks returns the N most recently completed tasks with their urge counts.
func (db *DB) GetRecentTasks(limit int) ([]models.TaskHistory, error) {
	rows, err := db.conn.Query(
		`SELECT t.id, t.name, t.started_at, t.duration_seconds,
		        (SELECT COUNT(*) FROM urges u WHERE u.task_id = t.id) AS urge_count
		 FROM tasks t
		 WHERE t.ended_at IS NOT NULL
		 ORDER BY t.started_at DESC
		 LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query recent tasks: %w", err)
	}
	defer rows.Close()

	var results []models.TaskHistory
	for rows.Next() {
		var h models.TaskHistory
		var startedAtStr string
		var durationSecs sql.NullInt64
		if err := rows.Scan(&h.ID, &h.Name, &startedAtStr, &durationSecs, &h.UrgeCount); err != nil {
			return nil, fmt.Errorf("scan task history: %w", err)
		}
		h.StartedAt, err = parseTime(startedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse started_at: %w", err)
		}
		if durationSecs.Valid {
			h.DurationSeconds = durationSecs.Int64
		}
		results = append(results, h)
	}
	return results, rows.Err()
}

// DeleteTask deletes a single task and its associated urges.
func (db *DB) DeleteTask(id int64) error {
	_, err := db.conn.Exec("DELETE FROM urges WHERE task_id = ?; DELETE FROM tasks WHERE id = ?;", id, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

// ResetAll deletes all tasks and urges, resetting the database to a fresh state.
func (db *DB) ResetAll() error {
	_, err := db.conn.Exec("DELETE FROM urges; DELETE FROM tasks;")
	if err != nil {
		return fmt.Errorf("reset database: %w", err)
	}
	return nil
}
