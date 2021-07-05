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

func (m *Match) String() string {
	m.HomeTeam.LoadOrganization()
	m.VisitingTeam.LoadOrganization()

	var str strings.Builder
	str.WriteString(m.Date.Format("Mon, Jan 02"))
	str.WriteString("\t")

	if m.Outcome.WinningTeam == nil {
		str.WriteString(m.HomeTeam.Organization.ShortName())
		str.WriteString(" ")
		str.WriteString(m.HomeTeam.ShortName())
		str.WriteString(" vs. ")
		str.WriteString(m.VisitingTeam.Organization.ShortName())
		str.WriteString(" ")
		str.WriteString(m.VisitingTeam.ShortName())
	} else {
		m.Outcome.WinningTeam.LoadOrganization()

		winner := m.Outcome.WinningTeam
		var loser *Team
		if winner == m.HomeTeam {
			loser = m.VisitingTeam
		} else {
			loser = m.HomeTeam
		}

		str.WriteString(winner.Organization.ShortName())
		str.WriteString(" ")
		str.WriteString(winner.ShortName())
		str.WriteString("   WON ")
		str.WriteString(fmt.Sprintf("%d - %d", m.Outcome.WinnerPoints, m.Outcome.LoserPoints))
		str.WriteString("   against ")
		str.WriteString(loser.Organization.ShortName())
		str.WriteString(" ")
		str.WriteString(loser.ShortName())
		str.WriteString(" ")
	}

	return str.String()
}
