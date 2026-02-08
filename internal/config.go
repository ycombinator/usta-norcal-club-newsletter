package internal

import (
	"time"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/formatters"
)

const asrcOrganizationID = 225

// Config holds the application configuration.
type Config struct {
	OrganizationID int
	TeamIDs        []int

	PastDuration   time.Duration
	FutureDuration time.Duration

	Formatter formatters.Formatter
}

// DefaultConfig returns the default application configuration.
func DefaultConfig() Config {
	return Config{
		OrganizationID: asrcOrganizationID,
		PastDuration:   10 * 24 * time.Hour,
		FutureDuration: 10 * 24 * time.Hour,
		Formatter: formatters.NewConsoleFormatter(),
	}
}
