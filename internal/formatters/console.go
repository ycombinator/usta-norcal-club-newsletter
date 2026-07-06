package formatters

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

type ConsoleFormatter struct{}

func NewConsoleFormatter() *ConsoleFormatter {
	return &ConsoleFormatter{}
}

func (c *ConsoleFormatter) FormatRecent(data *PreparedData, cfg Config) error {
	if !data.hasPastMatches() {
		return nil
	}

	var str strings.Builder
	str.WriteString("Recent matches:\n")
	table := tablewriter.NewWriter(&str)
	table.SetAutoWrapText(false)

	if data.DataFile != nil {
		for _, rec := range data.DataFile.PastMatches {
			table.Append([]string{
				dataFileDateDisplay(rec.Date),
				data.DataFile.OrgShortName + " " + rec.GenderEmoji + rec.Level + rec.Superscript,
				consoleOutcome(rec),
				consoleLocOpponent(rec.IsHome, rec.Opponent),
			})
		}
	} else {
		for _, am := range data.PastMatches {
			date, first, outcome, locOpponent := formatAnnotatedMatch(am, data.Org, data.OrgNames, cfg.Reader, cfg.Writer)
			table.Append([]string{date, first, outcome, locOpponent})
		}
	}

	table.Render()
	str.WriteString("\n")
	fmt.Fprint(cfg.Writer, str.String())
	return nil
}

func (c *ConsoleFormatter) FormatUpcoming(data *PreparedData, cfg Config) error {
	if !data.hasUpcomingMatches() {
		return nil
	}

	var str strings.Builder
	str.WriteString("Upcoming matches:\n")
	table := tablewriter.NewWriter(&str)
	table.SetAutoWrapText(false)

	if data.DataFile != nil {
		for _, rec := range data.DataFile.FutureMatches {
			date := dataFileDateDisplay(rec.Date)
			if rec.Time != "" {
				date += " " + dataFileMatchTime(rec.Time)
			}
			locOpponent := consoleLocOpponent(rec.IsHome, rec.Opponent)
			if rec.LocationNote != "" {
				locOpponent += fmt.Sprintf(" (at %s)", rec.LocationNote)
			}
			table.Append([]string{
				date,
				data.DataFile.OrgShortName + " " + rec.GenderEmoji + rec.Level + rec.Superscript,
				locOpponent,
			})
		}
	} else {
		for i, m := range data.FutureMatches {
			_, first, _, locOpponent := formatFutureMatch(m, data.Org, data.OrgNames, cfg.Reader, cfg.Writer)
			if loc, ok := data.LocationOverrides[i]; ok {
				locOpponent += fmt.Sprintf(" (at %s)", loc)
			}
			date := m.Date.Format("Mon, Jan 02 03:04 PM")
			table.Append([]string{date, first, locOpponent})
		}
	}

	table.Render()
	str.WriteString("\n")
	fmt.Fprint(cfg.Writer, str.String())
	return nil
}

func consoleOutcome(rec PastMatchRecord) string {
	var outcome string
	if rec.IsRainedOut {
		outcome = "rained out"
	} else if rec.IsIncomplete && rec.OutcomeText != "" {
		outcome = rec.OutcomeText + "*"
	} else if rec.IsIncomplete {
		outcome = "*"
	} else {
		outcome = rec.OutcomeText
	}
	switch rec.MatchType {
	case "playoff":
		outcome += " [playoff]"
	case "sectionals":
		outcome += " [Sectionals]"
	}
	return outcome
}

func consoleLocOpponent(isHome bool, opponent string) string {
	if isHome {
		return "vs. " + opponent
	}
	return "@ " + opponent
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
