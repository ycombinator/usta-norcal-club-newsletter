package formatters

import (
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
)

type PNGFormatter struct{}

func NewPNGFormatter() *PNGFormatter {
	return new(PNGFormatter)
}

func (p *PNGFormatter) Format(n *core.Newsletter, c Config) error {
	return nil
}
