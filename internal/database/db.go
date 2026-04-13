package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"avledger/internal/models"
	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB handle and exposes all CRUD operations.
type DB struct {
	conn *sql.DB
	Path string
}

// Open opens (or creates) the SQLite database at the standard data path.
func Open() (*DB, error) {
	dataDir, err := os.UserConfigDir()
	if err != nil {
		dataDir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine data directory: %w", err)
		}
	}

	dir := filepath.Join(dataDir, "avledger")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create data directory: %w", err)
	}

	path := filepath.Join(dir, "avledger.db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	db := &DB{conn: conn, Path: path}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return db, nil
}

// migrate creates the schema if it does not exist.
func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS log_entries (
			id                   INTEGER PRIMARY KEY AUTOINCREMENT,
			date                 TEXT    NOT NULL,
			aircraft_engine_type TEXT    NOT NULL,
			reg_marks            TEXT    NOT NULL,
			task_detail          TEXT    NOT NULL,
			category             TEXT    NOT NULL,
			job_type             TEXT    NOT NULL DEFAULT '',
			ata                  TEXT    NOT NULL DEFAULT '',
			work_order_number    TEXT    NOT NULL DEFAULT '',
			verified_by          TEXT    NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT ''
		);
	`)
	if err != nil {
		return err
	}

	// Safely add job_type column to existing tables
	var colName string
	err = db.conn.QueryRow("SELECT name FROM pragma_table_info('log_entries') WHERE name='job_type'").Scan(&colName)
	if err == sql.ErrNoRows {
		_, err = db.conn.Exec("ALTER TABLE log_entries ADD COLUMN job_type TEXT NOT NULL DEFAULT ''")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// ---- Log Entries ----

// ListEntries returns all log entries sorted by date ascending.
func (db *DB) ListEntries() ([]models.LogEntry, error) {
	rows, err := db.conn.Query(`
		SELECT id, date, aircraft_engine_type, reg_marks, task_detail,
		       category, job_type, ata, work_order_number, verified_by
		FROM log_entries
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LogEntry
	for rows.Next() {
		var e models.LogEntry
		if err := rows.Scan(
			&e.ID, &e.Date, &e.AircraftEngineType, &e.RegMarks, &e.TaskDetail,
			&e.Category, &e.JobType, &e.ATA, &e.WorkOrderNumber, &e.VerifiedBy,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if rows.Err() == nil {
		sortEntriesDesc(entries)
	}
	return entries, rows.Err()
}

// SearchEntries returns log entries that match the given query string.
func (db *DB) SearchEntries(query string) ([]models.LogEntry, error) {
	likeQuery := "%" + query + "%"
	rows, err := db.conn.Query(`
		SELECT id, date, aircraft_engine_type, reg_marks, task_detail,
		       category, job_type, ata, work_order_number, verified_by
		FROM log_entries
		WHERE task_detail LIKE ?
		   OR aircraft_engine_type LIKE ?
		   OR reg_marks LIKE ?
		   OR category LIKE ?
		   OR job_type LIKE ?
		   OR ata LIKE ?
		   OR work_order_number LIKE ?
		   OR verified_by LIKE ?
		ORDER BY id ASC
	`, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LogEntry
	for rows.Next() {
		var e models.LogEntry
		if err := rows.Scan(
			&e.ID, &e.Date, &e.AircraftEngineType, &e.RegMarks, &e.TaskDetail,
			&e.Category, &e.JobType, &e.ATA, &e.WorkOrderNumber, &e.VerifiedBy,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if rows.Err() == nil {
		sortEntriesDesc(entries)
	}
	return entries, rows.Err()
}

// CreateEntry inserts a new log entry and returns its assigned ID.
func (db *DB) CreateEntry(e models.LogEntry) (int64, error) {
	res, err := db.conn.Exec(`
		INSERT INTO log_entries
			(date, aircraft_engine_type, reg_marks, task_detail, category, job_type, ata, work_order_number, verified_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Date, e.AircraftEngineType, e.RegMarks, e.TaskDetail,
		e.Category, e.JobType, e.ATA, e.WorkOrderNumber, e.VerifiedBy,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateEntry updates an existing log entry by ID.
func (db *DB) UpdateEntry(e models.LogEntry) error {
	_, err := db.conn.Exec(`
		UPDATE log_entries SET
			date = ?, aircraft_engine_type = ?, reg_marks = ?,
			task_detail = ?, category = ?, job_type = ?, ata = ?,
			work_order_number = ?, verified_by = ?
		WHERE id = ?`,
		e.Date, e.AircraftEngineType, e.RegMarks, e.TaskDetail,
		e.Category, e.JobType, e.ATA, e.WorkOrderNumber, e.VerifiedBy, e.ID,
	)
	return err
}

// DeleteEntry removes a log entry by ID.
func (db *DB) DeleteEntry(id int64) error {
	_, err := db.conn.Exec(`DELETE FROM log_entries WHERE id = ?`, id)
	return err
}

// ---- Settings ----

// GetSetting retrieves a setting value by key, returning "" if not set.
func (db *DB) GetSetting(key string) (string, error) {
	var value string
	err := db.conn.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetSetting upserts a setting key/value pair.
func (db *DB) SetSetting(key, value string) error {
	_, err := db.conn.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

// ---- Sorting Helpers ----

func parseDate(d string) time.Time {
	d = strings.TrimSpace(d)
	formats := []string{
		"02/01/2006",
		"2/1/2006",
		"02/01/06",
		"2006-01-02",
		"02 Jan 2006",
		"2 Jan 2006",
		"02 JAN 2006",
		"2 JAN 2006",
		"02-01-2006",
		"2-1-2006",
		"02.01.2006",
		"2.1.2006",
		"02-Jan-2006",
		"2-Jan-2006",
		"Jan 02, 2006",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, d); err == nil {
			return t
		}
	}
	return time.Time{}
}

func sortEntriesDesc(entries []models.LogEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		t1 := parseDate(entries[i].Date)
		t2 := parseDate(entries[j].Date)
		if t1.Equal(t2) {
			return entries[i].ID > entries[j].ID
		}
		return t1.After(t2)
	})
}
