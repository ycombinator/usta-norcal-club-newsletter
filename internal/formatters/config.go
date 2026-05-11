package formatters

import (
	"io"
	"time"
)

// Config holds the formatter configuration.
type Config struct {
	OrganizationID int

	PastDuration   time.Duration
	FutureDuration time.Duration
	BoundaryDate   time.Time

	OutputDir string

	Reader io.Reader
	Writer io.Writer
}
