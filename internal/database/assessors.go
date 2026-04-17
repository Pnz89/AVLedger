package database

import (
	"avledger/internal/models"
)

// ListAssessors returns all assessors ordered by name.
func (db *DB) ListAssessors() ([]models.Assessor, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, license_number, company_approval
		FROM assessors
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assessors []models.Assessor
	for rows.Next() {
		var a models.Assessor
		if err := rows.Scan(&a.ID, &a.Name, &a.LicenseNumber, &a.CompanyApproval); err != nil {
			return nil, err
		}
		assessors = append(assessors, a)
	}
	return assessors, rows.Err()
}

// GetAssessorByName returns a single assessor by their exact name.
func (db *DB) GetAssessorByName(name string) (models.Assessor, error) {
	var a models.Assessor
	err := db.conn.QueryRow(`
		SELECT id, name, license_number, company_approval
		FROM assessors
		WHERE name = ?
	`, name).Scan(&a.ID, &a.Name, &a.LicenseNumber, &a.CompanyApproval)
	return a, err
}

// CreateAssessor inserts a new assessor.
func (db *DB) CreateAssessor(a models.Assessor) (int64, error) {
	res, err := db.conn.Exec(`
		INSERT INTO assessors (name, license_number, company_approval)
		VALUES (?, ?, ?)`,
		a.Name, a.LicenseNumber, a.CompanyApproval,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateAssessor updates an existing assessor.
func (db *DB) UpdateAssessor(a models.Assessor) error {
	_, err := db.conn.Exec(`
		UPDATE assessors SET
			name = ?, license_number = ?, company_approval = ?
		WHERE id = ?`,
		a.Name, a.LicenseNumber, a.CompanyApproval, a.ID,
	)
	return err
}

// DeleteAssessor Removes an assessor by ID.
func (db *DB) DeleteAssessor(id int64) error {
	_, err := db.conn.Exec(`DELETE FROM assessors WHERE id = ?`, id)
	return err
}
