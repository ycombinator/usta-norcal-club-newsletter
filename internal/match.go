package internal

import "time"

// Match represents a match consisting of multiple lines.
type Match struct {
	Date time.Time
	HomeTeam Team
	AwayTeam Team
	Location Organization
	
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