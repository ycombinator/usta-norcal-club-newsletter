package formatters

import (
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
)

type Formatter interface {
	Format(n *core.Newsletter, cfg Config) error
}
