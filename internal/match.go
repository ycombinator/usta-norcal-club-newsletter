package internal

import (
	"fmt"
	"strings"
	"time"
)

// Match represents a match consisting of multiple lines.
type Match struct {
	Date         time.Time
	HomeTeam     *Team
	VisitingTeam *Team
	Location     *Organization

	Outcome struct {
		WinningTeam  *Team
		WinnerPoints int
		LoserPoints  int
	}

	Lines []Line
}

// Line represents and individual line within a match.
type Line struct {
	HomePlayer1 Player
	HomePlayer2 Player

	AwayPlayer1 Player
	AwayPlayer2 Player

	WinnerScore string
	WinningTeam Team
}

func (m *Match) String(forOrg *Organization) string {
	m.HomeTeam.LoadOrganization()
	m.VisitingTeam.LoadOrganization()

	var firstTeam, secondTeam *Team
	var locator string
	if m.HomeTeam.Organization.Equals(forOrg) {
		firstTeam = m.HomeTeam
		secondTeam = m.VisitingTeam
		locator = "vs."
	} else {
		firstTeam = m.VisitingTeam
		secondTeam = m.HomeTeam
		locator = "@"
	}

	var str strings.Builder
	str.WriteString(m.Date.Format("Mon, Jan 02"))
	str.WriteString("\t")

	if m.Outcome.WinningTeam == nil {
		str.WriteString(firstTeam.Organization.ShortName())
		str.WriteString(" ")
		str.WriteString(firstTeam.ShortName())
		str.WriteString(" " + locator + " ")
		str.WriteString(secondTeam.Organization.ShortName())
		str.WriteString(" ")
		str.WriteString(secondTeam.ShortName())
	} else {
		m.Outcome.WinningTeam.LoadOrganization()

		var outcome string
		if m.Outcome.WinningTeam == firstTeam {
			outcome = fmt.Sprintf("WON %d - %d", m.Outcome.WinnerPoints, m.Outcome.LoserPoints)
		} else {
			outcome = fmt.Sprintf("LOST %d - %d", m.Outcome.LoserPoints, m.Outcome.WinnerPoints)
		}

		str.WriteString(firstTeam.Organization.ShortName())
		str.WriteString(" ")
		str.WriteString(firstTeam.ShortName())
		str.WriteString("   " + outcome + "  ")
		str.WriteString(locator + " ")
		str.WriteString(secondTeam.Organization.ShortName())
		str.WriteString(" ")
		str.WriteString(secondTeam.ShortName())
		str.WriteString(" ")
	}

	return str.String()
}
