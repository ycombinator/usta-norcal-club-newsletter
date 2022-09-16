package formatters

import "time"

// Config holds the formatter configuration.
type Config struct {
	OrganizationID int

	PastDuration   time.Duration
	FutureDuration time.Duration
}
