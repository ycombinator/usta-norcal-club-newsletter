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

	OutputDir    string
	DataFilePath string // path to intermediate JSON data file; loaded if found, saved otherwise

	Reader io.Reader
	Writer io.Writer
}
