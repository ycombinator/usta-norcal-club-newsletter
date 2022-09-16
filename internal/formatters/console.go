package formatters

import (
	"fmt"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal"
)

type ConsoleFormatter struct{}

func (c *ConsoleFormatter) Format(n *internal.Newsletter) error {
	fmt.Println(n.String())
	return nil
}
