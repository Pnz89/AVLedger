package models

// Aircraft represents a saved aircraft with its registration, aircraft type, and engine type.
type Aircraft struct {
	ID           int64
	Registration string
	Aircraft     string
	Engine       string
}
