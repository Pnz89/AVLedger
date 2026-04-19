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
	"io"
)

// DB wraps the sql.DB handle and exposes all CRUD operations.
type DB struct {
	conn *sql.DB
	Path string
}

// Open opens (or creates) the SQLite database at the specified path. If empty, defaults to standard config dir.
func Open(customPath string) (*DB, error) {
	var path, dir string

	if customPath != "" {
		path = customPath
		dir = filepath.Dir(path)
	} else {
		dataDir, err := os.UserConfigDir()
		if err != nil {
			dataDir, err = os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("cannot determine data directory: %w", err)
			}
		}
		dir = filepath.Join(dataDir, "avledger")
		path = filepath.Join(dir, "avledger.db")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create data directory: %w", err)
	}

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

// MoveTo safely closes the active connection, physically moves the database to the target directory,
// reopens the connection, and updates the db context struct seamlessly.
func (db *DB) MoveTo(newDir string) error {
	if err := db.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(newDir, 0755); err != nil {
		return err
	}

	newPath := filepath.Join(newDir, "avledger.db")

	// Prevent overwriting an existing database in the destination folder
	if _, err := os.Stat(newPath); err == nil {
		// Attempt fallback reopen
		db.conn, _ = sql.Open("sqlite", db.Path)
		return fmt.Errorf("a database already exists at the destination: %s", newPath)
	}

	if err := moveFile(db.Path, newPath); err != nil {
		// attempt fallback reopen
		db.conn, _ = sql.Open("sqlite", db.Path)
		return err
	}

	conn, err := sql.Open("sqlite", newPath)
	if err != nil {
		return err
	}
	db.conn = conn
	db.Path = newPath

	return nil
}

// SwitchTo safely closes the active connection and reopens it pointing to the specified new path,
// without moving any files.
func (db *DB) SwitchTo(newPath string) error {
	if err := db.Close(); err != nil {
		// Proceed anyway just in case
	}

	conn, err := sql.Open("sqlite", newPath)
	if err != nil {
		// Attempt fallback reopen
		db.conn, _ = sql.Open("sqlite", db.Path)
		return err
	}
	
	db.conn = conn
	db.Path = newPath

	return nil
}

// moveFile bridges file relocation smoothly bridging partitions.
func moveFile(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	// Fallback to copy+delete if os.Rename fails (e.g. across drives)
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	out.Sync()

	in.Close() // Must close explicitly before remove
	return os.Remove(src)
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
			duration             TEXT    NOT NULL DEFAULT '',
			verified_by          TEXT    NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS assessors (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			name             TEXT    NOT NULL UNIQUE,
			license_number   TEXT    NOT NULL DEFAULT '',
			company_approval TEXT    NOT NULL DEFAULT ''
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

	// Safely add duration column to existing tables
	err = db.conn.QueryRow("SELECT name FROM pragma_table_info('log_entries') WHERE name='duration'").Scan(&colName)
	if err == sql.ErrNoRows {
		_, err = db.conn.Exec("ALTER TABLE log_entries ADD COLUMN duration TEXT NOT NULL DEFAULT ''")
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
		       category, job_type, ata, work_order_number, duration, verified_by
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
			&e.Category, &e.JobType, &e.ATA, &e.WorkOrderNumber, &e.Duration, &e.VerifiedBy,
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

// SearchEntries returns log entries that match the given filter options.
func (db *DB) SearchEntries(f models.FilterOptions) ([]models.LogEntry, error) {
	queryBuilder := `
		SELECT id, date, aircraft_engine_type, reg_marks, task_detail,
		       category, job_type, ata, work_order_number, duration, verified_by
		FROM log_entries
		WHERE 1=1`

	var args []interface{}

	if f.SearchQuery != "" {
		likeQuery := "%" + f.SearchQuery + "%"
		queryBuilder += ` AND (task_detail LIKE ?
		   OR aircraft_engine_type LIKE ?
		   OR reg_marks LIKE ?
		   OR category LIKE ?
		   OR job_type LIKE ?
		   OR ata LIKE ?
		   OR work_order_number LIKE ?
		   OR verified_by LIKE ?)`
		args = append(args, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery, likeQuery)
	}

	if f.AircraftEngineType != "" {
		queryBuilder += ` AND aircraft_engine_type = ?`
		args = append(args, f.AircraftEngineType)
	}
	if f.RegMarks != "" {
		queryBuilder += ` AND reg_marks = ?`
		args = append(args, f.RegMarks)
	}
	if f.Category != "" {
		queryBuilder += ` AND category = ?`
		args = append(args, f.Category)
	}
	if f.JobType != "" {
		queryBuilder += ` AND job_type = ?`
		args = append(args, f.JobType)
	}

	queryBuilder += ` ORDER BY id ASC`

	rows, err := db.conn.Query(queryBuilder, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.LogEntry
	for rows.Next() {
		var e models.LogEntry
		if err := rows.Scan(
			&e.ID, &e.Date, &e.AircraftEngineType, &e.RegMarks, &e.TaskDetail,
			&e.Category, &e.JobType, &e.ATA, &e.WorkOrderNumber, &e.Duration, &e.VerifiedBy,
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
			(date, aircraft_engine_type, reg_marks, task_detail, category, job_type, ata, work_order_number, duration, verified_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Date, e.AircraftEngineType, e.RegMarks, e.TaskDetail,
		e.Category, e.JobType, e.ATA, e.WorkOrderNumber, e.Duration, e.VerifiedBy,
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
			work_order_number = ?, duration = ?, verified_by = ?
		WHERE id = ?`,
		e.Date, e.AircraftEngineType, e.RegMarks, e.TaskDetail,
		e.Category, e.JobType, e.ATA, e.WorkOrderNumber, e.Duration, e.VerifiedBy, e.ID,
	)
	return err
}

// DeleteEntry removes a log entry by ID.
func (db *DB) DeleteEntry(id int64) error {
	_, err := db.conn.Exec(`DELETE FROM log_entries WHERE id = ?`, id)
	return err
}

// GetDistinctValues retrieves a sorted list of unique entries for a specified column.
func (db *DB) GetDistinctValues(column string) ([]string, error) {
	validCols := map[string]bool{
		"aircraft_engine_type": true,
		"reg_marks":            true,
		"category":             true,
		"job_type":             true,
	}
	if !validCols[column] {
		return nil, fmt.Errorf("invalid column: %s", column)
	}

	query := fmt.Sprintf("SELECT DISTINCT %s FROM log_entries WHERE %s != '' ORDER BY %s ASC", column, column, column)
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	return values, rows.Err()
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
