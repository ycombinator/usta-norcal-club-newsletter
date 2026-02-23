package usta

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func (m *Match) ForOrganization(forOrg *Organization) (date time.Time, first string, outcome string, locator string, second string) {
	// Formatting runs after all data is loaded; use Background context since
	// organizations will already be cached and no new HTTP calls are expected.
	ctx := context.Background()
	m.HomeTeam.LoadOrganization(ctx)
	m.VisitingTeam.LoadOrganization(ctx)

	var firstTeam, secondTeam *Team
	if m.HomeTeam.Organization.Equals(forOrg) {
		firstTeam = m.HomeTeam
		secondTeam = m.VisitingTeam
		locator = "vs."
	} else {
		firstTeam = m.VisitingTeam
		secondTeam = m.HomeTeam
		locator = "@"
	}

	date = m.Date

	first = firstTeam.Organization.ShortName() + " " + firstTeam.ShortName()
	second = cases.Title(language.English).String(strings.ToLower(secondTeam.Organization.Name))
	if m.Outcome.WinningTeam != nil {
		m.Outcome.WinningTeam.LoadOrganization(ctx)

		if m.Outcome.WinningTeam == firstTeam {
			outcome = fmt.Sprintf("won %d - %d", m.Outcome.WinnerPoints, m.Outcome.LoserPoints)
		} else {
			outcome = fmt.Sprintf("lost %d - %d", m.Outcome.LoserPoints, m.Outcome.WinnerPoints)
		}
	}

	return
}
