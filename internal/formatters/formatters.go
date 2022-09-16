package formatters

import "github.com/ycombinator/usta-norcal-club-newsletter/internal"

type Formatter interface {
	Format(n internal.Newsletter) error
}
