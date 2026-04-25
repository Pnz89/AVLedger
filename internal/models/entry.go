package models

// LogEntry represents a single maintenance logbook entry,
// matching the fields of the RPPL.F.054B Section 5 form.
type LogEntry struct {
	ID                 int64
	Date               string // DD/MM/YYYY
	AircraftEngineType string // e.g. "B737 NG (CFM56)"
	RegMarks           string // e.g. "EI-DAZ"
	TaskDetail         string // Full description of the maintenance task
	Category           string // A1, A2, A3, A4, B1.1, B1.2, B1.3, B1.4, B2, B3, C, Mech
	JobType            string // e.g. Line, Base, Mod
	ATA                string // ATA chapter code
	WorkOrderNumber    string
	Duration           string // Task duration in hours, e.g. "2.5"
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
	StartDate          string
	EndDate            string
	Category           string
	JobType            string
	ATA                string
}

// UserProfile represents a user and their database location.
type UserProfile struct {
	Name   string `json:"name"`
	DBPath string `json:"dbPath"`
}
