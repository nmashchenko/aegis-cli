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

// migrations is an ordered list of schema-only migration functions.
// Each function receives a transaction and returns an error.
// Append new migrations to the end; never reorder or remove existing entries.
var migrations = []func(tx *sql.Tx) error{
	// 1: initial tables
	func(tx *sql.Tx) error {
		_, err := tx.Exec(`
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

			CREATE TABLE IF NOT EXISTS papers (
				id         INTEGER PRIMARY KEY AUTOINCREMENT,
				mood_tier  TEXT NOT NULL,
				title      TEXT NOT NULL,
				url        TEXT NOT NULL,
				highlight  TEXT NOT NULL
			);
		`)
		return err
	},
	// 2: add pause columns to tasks
	func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			ALTER TABLE tasks ADD COLUMN paused_at DATETIME;
			ALTER TABLE tasks ADD COLUMN total_paused_seconds INTEGER NOT NULL DEFAULT 0;
		`)
		return err
	},
	// future migrations append here
}

func (db *DB) migrate() error {
	if err := db.runMigrations(); err != nil {
		return err
	}

	// Seed papers if table is empty (independent of migrations)
	var count int
	if err := db.conn.QueryRow("SELECT COUNT(*) FROM papers").Scan(&count); err != nil {
		return fmt.Errorf("check papers count: %w", err)
	}
	if count == 0 {
		if err := db.seedPapers(); err != nil {
			return fmt.Errorf("seed papers: %w", err)
		}
	}

	return nil
}

// runMigrations bootstraps the schema_migrations table and applies any
// pending migrations in order. Each migration runs inside its own transaction.
func (db *DB) runMigrations() error {
	_, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	for i, fn := range migrations {
		version := i + 1

		var exists int
		err := db.conn.QueryRow("SELECT 1 FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err == nil {
			continue // already applied
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration %d: %w", version, err)
		}

		tx, err := db.conn.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", version, err)
		}

		if err := fn(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("run migration %d: %w", version, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", version, err)
		}
	}

	return nil
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
		"SELECT id, name, started_at, ended_at, duration_seconds, limit_seconds, paused_at, total_paused_seconds FROM tasks WHERE ended_at IS NULL LIMIT 1",
	)

	var task models.Task
	var startedAt string
	var endedAt sql.NullString
	var durationSeconds sql.NullInt64
	var limitSeconds sql.NullInt64
	var pausedAt sql.NullString
	var totalPausedSeconds int64

	err := row.Scan(&task.ID, &task.Name, &startedAt, &endedAt, &durationSeconds, &limitSeconds, &pausedAt, &totalPausedSeconds)
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

	if pausedAt.Valid {
		t, err := parseTime(pausedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse paused_at: %w", err)
		}
		task.PausedAt = &t
	}

	task.TotalPausedSeconds = totalPausedSeconds

	return &task, nil
}

func (db *DB) PauseTask(id int64) error {
	_, err := db.conn.Exec(
		"UPDATE tasks SET paused_at = ? WHERE id = ? AND ended_at IS NULL AND paused_at IS NULL",
		time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("pause task: %w", err)
	}
	return nil
}

func (db *DB) ResumeTask(id int64) error {
	var pausedAtStr string
	err := db.conn.QueryRow("SELECT paused_at FROM tasks WHERE id = ? AND paused_at IS NOT NULL", id).Scan(&pausedAtStr)
	if err != nil {
		return fmt.Errorf("get paused_at: %w", err)
	}

	pausedAt, err := parseTime(pausedAtStr)
	if err != nil {
		return fmt.Errorf("parse paused_at: %w", err)
	}

	pausedSecs := int64(time.Since(pausedAt).Seconds())

	_, err = db.conn.Exec(
		"UPDATE tasks SET paused_at = NULL, total_paused_seconds = total_paused_seconds + ? WHERE id = ?",
		pausedSecs, id,
	)
	if err != nil {
		return fmt.Errorf("resume task: %w", err)
	}
	return nil
}

func (db *DB) StopTask(id int64) error {
	var startedAtStr string
	var pausedAtStr sql.NullString
	var totalPausedSeconds int64
	err := db.conn.QueryRow(
		"SELECT started_at, paused_at, total_paused_seconds FROM tasks WHERE id = ? AND ended_at IS NULL",
		id,
	).Scan(&startedAtStr, &pausedAtStr, &totalPausedSeconds)
	if err != nil {
		return fmt.Errorf("get task times: %w", err)
	}

	startedAt, err := parseTime(startedAtStr)
	if err != nil {
		return fmt.Errorf("parse started_at: %w", err)
	}

	now := time.Now().UTC()

	// If paused, accumulate the final pause chunk
	if pausedAtStr.Valid {
		pausedAt, err := parseTime(pausedAtStr.String)
		if err != nil {
			return fmt.Errorf("parse paused_at: %w", err)
		}
		totalPausedSeconds += int64(now.Sub(pausedAt).Seconds())
	}

	durationSecs := int64(now.Sub(startedAt).Seconds()) - totalPausedSeconds

	_, err = db.conn.Exec(
		"UPDATE tasks SET ended_at = ?, duration_seconds = ?, paused_at = NULL, total_paused_seconds = ? WHERE id = ? AND ended_at IS NULL",
		now, durationSecs, totalPausedSeconds, id,
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

// GetDailyUrgeCounts returns urge counts per day between start and end, filling gaps with zero.
// Groups in Go because SQLite's DATE() cannot parse the timestamp format modernc.org/sqlite produces.
func (db *DB) GetDailyUrgeCounts(start, end time.Time) ([]models.DailyUrgeCount, error) {
	rows, err := db.conn.Query(
		`SELECT timestamp FROM urges
		 WHERE timestamp >= ? AND timestamp < ?`,
		start.UTC(), end.UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("query urge timestamps: %w", err)
	}
	defer rows.Close()

	countMap := make(map[string]int)
	for rows.Next() {
		var tsStr string
		if err := rows.Scan(&tsStr); err != nil {
			return nil, fmt.Errorf("scan urge timestamp: %w", err)
		}
		t, err := parseTime(tsStr)
		if err != nil {
			continue
		}
		key := t.Local().Format("2006-01-02")
		countMap[key]++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fill in all days in range
	var results []models.DailyUrgeCount
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
	for d := startDay; d.Before(endDay); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		results = append(results, models.DailyUrgeCount{
			Date:  d,
			Count: countMap[key],
		})
	}

	return results, nil
}

// GetDailyTaskCounts returns completed task counts per day between start and end, filling gaps with zero.
// Groups in Go because SQLite's DATE() cannot parse the timestamp format modernc.org/sqlite produces.
func (db *DB) GetDailyTaskCounts(start, end time.Time) ([]models.DailyUrgeCount, error) {
	rows, err := db.conn.Query(
		`SELECT started_at FROM tasks
		 WHERE ended_at IS NOT NULL
		   AND started_at >= ? AND started_at < ?`,
		start.UTC(), end.UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("query task timestamps: %w", err)
	}
	defer rows.Close()

	countMap := make(map[string]int)
	for rows.Next() {
		var tsStr string
		if err := rows.Scan(&tsStr); err != nil {
			return nil, fmt.Errorf("scan task timestamp: %w", err)
		}
		t, err := parseTime(tsStr)
		if err != nil {
			continue
		}
		key := t.Local().Format("2006-01-02")
		countMap[key]++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var results []models.DailyUrgeCount
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
	for d := startDay; d.Before(endDay); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		results = append(results, models.DailyUrgeCount{
			Date:  d,
			Count: countMap[key],
		})
	}

	return results, nil
}

// GetRandomHighlight returns a random paper highlight for the given mood tier.
func (db *DB) GetRandomHighlight(moodTier string) (*models.PaperHighlight, error) {
	row := db.conn.QueryRow(
		"SELECT title, url, highlight FROM papers WHERE mood_tier = ? ORDER BY RANDOM() LIMIT 1",
		moodTier,
	)
	var h models.PaperHighlight
	err := row.Scan(&h.Title, &h.URL, &h.Highlight)
	if err != nil {
		return nil, fmt.Errorf("get random highlight: %w", err)
	}
	return &h, nil
}

// ResetAll deletes all tasks and urges, resetting the database to a fresh state.
func (db *DB) ResetAll() error {
	_, err := db.conn.Exec("DELETE FROM urges; DELETE FROM tasks;")
	if err != nil {
		return fmt.Errorf("reset database: %w", err)
	}
	return nil
}
