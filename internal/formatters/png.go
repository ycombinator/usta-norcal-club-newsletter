package formatters

import (
	"github.com/ycombinator/usta-norcal-club-newsletter/internal"
)

type PNGFormatter struct{}

func (p *PNGFormatter) Format(n *internal.Newsletter) error {
	return nil
}
