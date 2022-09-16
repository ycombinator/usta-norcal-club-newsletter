package formatters

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal"
)

type ConsoleFormatter struct{}

func (c *ConsoleFormatter) Format(n *internal.Newsletter) error {
	var pastMatches, futureMatches []internal.Match
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	cfg := n.Config()
	start := now.Add(-1 * cfg.PastDuration)
	end := now.Add(cfg.FutureDuration)

	org := n.Organization()
	for _, t := range org.Teams {
		for _, m := range t.Matches {
			if (m.Date.Equal(now) || m.Date.After(now)) && m.Date.Before(end) {
				futureMatches = append(futureMatches, m)
			} else if m.Date.Before(now) && (m.Date.Equal(start) || m.Date.After(start)) {
				pastMatches = append(pastMatches, m)
			}
		}
	}

	// Sort matches
	sort.Slice(pastMatches, func(i, j int) bool {
		return pastMatches[i].Date.Before(pastMatches[j].Date)
	})
	sort.Slice(futureMatches, func(i, j int) bool {
		return futureMatches[i].Date.Before(futureMatches[j].Date)
	})

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
				date.Format("Mon, Jan 02"),
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
