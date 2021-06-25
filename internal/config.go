package internal

import (
	"time"
)

const avacOrganizationID = 226

// Config holds the application configuration.
type Config struct {
	OrganizationID int
	PastDuration time.Duration
	FutureDuration time.Duration
}

// DefaultConfig returns the default application configuration.
func DefaultConfig() Config {
	return Config {
		OrganizationID: avacOrganizationID,
		PastDuration: 7 * 24 * time.Hour,
		FutureDuration: 7 * 24 * time.Hour,
	}
}