package database

import (
	"avledger/internal/models"
)

// ListAircrafts returns all aircrafts ordered by registration.
func (db *DB) ListAircrafts() ([]models.Aircraft, error) {
	rows, err := db.conn.Query(`
		SELECT id, registration, aircraft, engine
		FROM aircrafts
		ORDER BY registration ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aircrafts []models.Aircraft
	for rows.Next() {
		var a models.Aircraft
		if err := rows.Scan(&a.ID, &a.Registration, &a.Aircraft, &a.Engine); err != nil {
			return nil, err
		}
		aircrafts = append(aircrafts, a)
	}
	return aircrafts, rows.Err()
}

// GetAircraftByRegistration returns a single aircraft by its exact registration.
func (db *DB) GetAircraftByRegistration(registration string) (models.Aircraft, error) {
	var a models.Aircraft
	err := db.conn.QueryRow(`
		SELECT id, registration, aircraft, engine
		FROM aircrafts
		WHERE registration = ?
	`, registration).Scan(&a.ID, &a.Registration, &a.Aircraft, &a.Engine)
	return a, err
}

// CreateAircraft inserts a new aircraft.
func (db *DB) CreateAircraft(a models.Aircraft) (int64, error) {
	res, err := db.conn.Exec(`
		INSERT INTO aircrafts (registration, aircraft, engine)
		VALUES (?, ?, ?)`,
		a.Registration, a.Aircraft, a.Engine,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateAircraft updates an existing aircraft.
func (db *DB) UpdateAircraft(a models.Aircraft) error {
	_, err := db.conn.Exec(`
		UPDATE aircrafts SET
			registration = ?, aircraft = ?, engine = ?
		WHERE id = ?`,
		a.Registration, a.Aircraft, a.Engine, a.ID,
	)
	return err
}

// DeleteAircraft Removes an aircraft by ID.
func (db *DB) DeleteAircraft(id int64) error {
	_, err := db.conn.Exec(`DELETE FROM aircrafts WHERE id = ?`, id)
	return err
}
