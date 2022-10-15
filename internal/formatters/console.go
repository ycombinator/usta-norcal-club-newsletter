package formatters

import (
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
)

type ConsoleFormatter struct{}

func NewConsoleFormatter() *ConsoleFormatter {
	return new(ConsoleFormatter)
}

func (c *ConsoleFormatter) Format(n *core.Newsletter, cfg Config) error {
	org := n.Organization()
	pastMatches, futureMatches := org.Matches(cfg.PastDuration, cfg.FutureDuration)

	var str strings.Builder

	if len(pastMatches) > 0 {
		str.WriteString("Recent matches:\n")
		table := tablewriter.NewWriter(&str)
		table.SetAutoWrapText(false)
		for _, m := range pastMatches {
			date, first, outcome, locator, second := m.ForOrganization(org)
			table.Append([]string{
				date.Format("Mon, Jan 02"),
				first,
				outcome,
				locator + " " + second,
			})
		}
		table.Render()
		str.WriteString("\n")
	}

	if len(futureMatches) > 0 {
		str.WriteString("Upcoming matches:\n")
		table := tablewriter.NewWriter(&str)
		table.SetAutoWrapText(false)
		for _, m := range futureMatches {
			date, first, _, locator, second := m.ForOrganization(org)
			table.Append([]string{
				date.Format("Mon, Jan 02 03:04 PM"),
				first,
				locator + " " + second,
			})
		}
		table.Render()
		str.WriteString("\n")
	}

	fmt.Println(str.String())
	return nil
}
