package formatters

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

type ConsoleFormatter struct {
	reader io.Reader
	writer io.Writer
}

func NewConsoleFormatter() *ConsoleFormatter {
	return &ConsoleFormatter{reader: os.Stdin, writer: os.Stdout}
}

func (c *ConsoleFormatter) Format(n *core.Newsletter, cfg Config) error {
	data, err := Prepare(n, cfg, c.reader, c.writer)
	if err != nil {
		return err
	}

	var str strings.Builder

	if len(data.PastMatches) > 0 {
		str.WriteString("Recent matches:\n")
		table := tablewriter.NewWriter(&str)
		table.SetAutoWrapText(false)
		for _, am := range data.PastMatches {
			date, first, outcome, locOpponent := formatAnnotatedMatch(am, data.Org, data.OrgNames, c.reader, c.writer)
			table.Append([]string{
				date,
				first,
				outcome,
				locOpponent,
			})
		}
		table.Render()
		str.WriteString("\n")
	}

	if len(data.FutureMatches) > 0 {
		str.WriteString("Upcoming matches:\n")
		table := tablewriter.NewWriter(&str)
		table.SetAutoWrapText(false)
		for i, m := range data.FutureMatches {
			_, first, _, locOpponent := formatFutureMatch(m, data.Org, data.OrgNames, c.reader, c.writer)
			if loc, ok := data.LocationOverrides[i]; ok {
				locOpponent += fmt.Sprintf(" (at %s)", loc)
			}
			date := m.Date.Format("Mon, Jan 02 03:04 PM")
			table.Append([]string{
				date,
				first,
				locOpponent,
			})
		}
		table.Render()
		str.WriteString("\n")
	}

	fmt.Fprint(c.writer, str.String())

	return data.Save()
}

func formatAnnotatedMatch(am AnnotatedMatch, org *usta.Organization, names *OrgNames, reader io.Reader, writer io.Writer) (date, first, outcome, locOpponent string) {
	m := am.Match
	ourTeam, opponent, isHome := resolveTeams(m, org)
	opponent.LoadOrganization(context.Background())

	date = m.Date.Format("Mon, Jan 02")
	first = ourTeam.Organization.ShortName() + " " + ourTeam.ShortName()
	opName := opponentDisplayName(names, reader, writer, opponent.Organization)

	locator := "vs."
	if !isHome {
		locator = "@"
	}
	locOpponent = locator + " " + opName

	if am.Annotation.RainedOut {
		outcome = "rained out"
	} else if am.Annotation.Score != "" {
		outcome = am.Annotation.Score + "*"
	} else if am.Annotation.Footnote != "" {
		outcome = "*"
	} else if m.Outcome.WinningTeam != nil {
		m.Outcome.WinningTeam.LoadOrganization(context.Background())
		if m.Outcome.WinningTeam.Organization.Equals(ourTeam.Organization) || m.Outcome.WinningTeam == ourTeam {
			outcome = fmt.Sprintf("won %d - %d", m.Outcome.WinnerPoints, m.Outcome.LoserPoints)
		} else {
			outcome = fmt.Sprintf("lost %d - %d", m.Outcome.LoserPoints, m.Outcome.WinnerPoints)
		}
	}

	if am.Annotation.MatchType == Playoff {
		outcome += " [playoff]"
	} else if am.Annotation.MatchType == Sectionals {
		outcome += " [Sectionals]"
	}

	return
}

func formatFutureMatch(m usta.Match, org *usta.Organization, names *OrgNames, reader io.Reader, writer io.Writer) (date, first, outcome, locOpponent string) {
	ourTeam, opponent, isHome := resolveTeams(m, org)
	opponent.LoadOrganization(context.Background())

	date = m.Date.Format("Mon, Jan 02 03:04 PM")
	first = ourTeam.Organization.ShortName() + " " + ourTeam.ShortName()
	opName := opponentDisplayName(names, reader, writer, opponent.Organization)

	locator := "vs."
	if !isHome {
		locator = "@"
	}
	locOpponent = locator + " " + opName

	return
}
