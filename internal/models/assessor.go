package models

// Assessor represents a person who can verify maintenance tasks.
type Assessor struct {
	ID              int64
	Name            string
	LicenseNumber   string
	CompanyApproval string
}
