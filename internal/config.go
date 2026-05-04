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

	RecentFormatter   formatters.RecentFormatter
	UpcomingFormatter formatters.UpcomingFormatter
}

// DefaultConfig returns the default application configuration.
func DefaultConfig() Config {
	f := formatters.NewJPEGFormatter()
	return Config{
		OrganizationID:    asrcOrganizationID,
		PastDuration:      7 * 24 * time.Hour,
		FutureDuration:    7 * 24 * time.Hour,
		RecentFormatter:   f,
		UpcomingFormatter: f,
	}
}
