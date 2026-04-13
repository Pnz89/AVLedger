package models

// LogEntry represents a single maintenance logbook entry,
// matching the fields of the RPPL.F.054B Section 5 form.
type LogEntry struct {
	ID                 int64
	Date               string // DD/MM/YYYY
	AircraftEngineType string // e.g. "B737 NG (CFM56)"
	RegMarks           string // e.g. "EI-DAZ"
	TaskDetail         string // Full description of the maintenance task
	Category           string // A, B1, B2, or C
	JobType            string // e.g. Line, Base, Mod
	ATA                string // ATA chapter code
	WorkOrderNumber    string
	VerifiedBy         string // Name + authorisation number / AML number
}

// Settings holds user-configurable application settings.
type Settings struct {
	HolderName    string // Logbook holder full name
	LicenceNumber string // Licence / AML number for PDF footer
}

// FilterOptions holds user-selected filters for searching entries.
type FilterOptions struct {
	SearchQuery        string
	AircraftEngineType string
	RegMarks           string
	Category           string
	JobType            string
}
