package usta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestForOrganization(t *testing.T) {
	asrc := &Organization{ID: 225, Name: "Almaden Swim Racquet Club"}
	avac := &Organization{ID: 300, Name: "Almaden Valley Athletic Club"}
	lgsrc := &Organization{ID: 400, Name: "Los Gatos Swim Racquet Club"}

	asrcTeam := &Team{ID: 1, Name: "Adult 18+ Womens 3.5", Organization: asrc}
	avacTeam := &Team{ID: 2, Name: "Adult 18+ Womens 3.5", Organization: avac}
	lgsrcTeam := &Team{ID: 3, Name: "Adult 18+ Mens 4.0", Organization: lgsrc, Extra: true}

	matchDate := time.Date(2026, 4, 23, 19, 30, 0, 0, time.Local)

	t.Run("org team home win", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     asrcTeam,
			VisitingTeam: avacTeam,
		}
		m.Outcome.WinningTeam = asrcTeam
		m.Outcome.WinnerPoints = 3
		m.Outcome.LoserPoints = 2

		date, first, outcome, locator, second := m.ForOrganization(asrc)
		require.Equal(t, matchDate, date)
		require.Contains(t, first, "ASRC")
		require.Equal(t, "won 3 - 2", outcome)
		require.Equal(t, "vs.", locator)
		require.Equal(t, "Almaden Valley Athletic Club", second)
	})

	t.Run("org team away loss", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     avacTeam,
			VisitingTeam: asrcTeam,
		}
		m.Outcome.WinningTeam = avacTeam
		m.Outcome.WinnerPoints = 3
		m.Outcome.LoserPoints = 2

		date, first, outcome, locator, second := m.ForOrganization(asrc)
		require.Equal(t, matchDate, date)
		require.Contains(t, first, "ASRC")
		require.Equal(t, "lost 2 - 3", outcome)
		require.Equal(t, "@", locator)
		require.Equal(t, "Almaden Valley Athletic Club", second)
	})

	t.Run("org team home loss", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     asrcTeam,
			VisitingTeam: avacTeam,
		}
		m.Outcome.WinningTeam = avacTeam
		m.Outcome.WinnerPoints = 4
		m.Outcome.LoserPoints = 1

		_, first, outcome, locator, second := m.ForOrganization(asrc)
		require.Contains(t, first, "ASRC")
		require.Equal(t, "lost 1 - 4", outcome)
		require.Equal(t, "vs.", locator)
		require.Equal(t, "Almaden Valley Athletic Club", second)
	})

	t.Run("org team away win", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     avacTeam,
			VisitingTeam: asrcTeam,
		}
		m.Outcome.WinningTeam = asrcTeam
		m.Outcome.WinnerPoints = 3
		m.Outcome.LoserPoints = 2

		_, first, outcome, locator, _ := m.ForOrganization(asrc)
		require.Contains(t, first, "ASRC")
		require.Equal(t, "won 3 - 2", outcome)
		require.Equal(t, "@", locator)
	})

	t.Run("extra team away loss", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     avacTeam,
			VisitingTeam: lgsrcTeam,
		}
		m.Outcome.WinningTeam = avacTeam
		m.Outcome.WinnerPoints = 3
		m.Outcome.LoserPoints = 2

		_, first, outcome, locator, second := m.ForOrganization(asrc)
		require.Contains(t, first, "LGSRC")
		require.Equal(t, "lost 2 - 3", outcome)
		require.Equal(t, "@", locator)
		require.Equal(t, "Almaden Valley Athletic Club", second)
	})

	t.Run("extra team home win", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     lgsrcTeam,
			VisitingTeam: avacTeam,
		}
		m.Outcome.WinningTeam = lgsrcTeam
		m.Outcome.WinnerPoints = 4
		m.Outcome.LoserPoints = 1

		_, first, outcome, locator, second := m.ForOrganization(asrc)
		require.Contains(t, first, "LGSRC")
		require.Equal(t, "won 4 - 1", outcome)
		require.Equal(t, "vs.", locator)
		require.Equal(t, "Almaden Valley Athletic Club", second)
	})

	t.Run("extra team home loss", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     lgsrcTeam,
			VisitingTeam: avacTeam,
		}
		m.Outcome.WinningTeam = avacTeam
		m.Outcome.WinnerPoints = 3
		m.Outcome.LoserPoints = 2

		_, first, outcome, locator, _ := m.ForOrganization(asrc)
		require.Contains(t, first, "LGSRC")
		require.Equal(t, "lost 2 - 3", outcome)
		require.Equal(t, "vs.", locator)
	})

	t.Run("extra team away win", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     avacTeam,
			VisitingTeam: lgsrcTeam,
		}
		m.Outcome.WinningTeam = lgsrcTeam
		m.Outcome.WinnerPoints = 5
		m.Outcome.LoserPoints = 0

		_, first, outcome, locator, _ := m.ForOrganization(asrc)
		require.Contains(t, first, "LGSRC")
		require.Equal(t, "won 5 - 0", outcome)
		require.Equal(t, "@", locator)
	})

	t.Run("upcoming match no outcome", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     asrcTeam,
			VisitingTeam: avacTeam,
		}

		_, first, outcome, locator, second := m.ForOrganization(asrc)
		require.Contains(t, first, "ASRC")
		require.Equal(t, "", outcome)
		require.Equal(t, "vs.", locator)
		require.Equal(t, "Almaden Valley Athletic Club", second)
	})

	t.Run("extra team upcoming match no outcome", func(t *testing.T) {
		m := Match{
			Date:         matchDate,
			HomeTeam:     avacTeam,
			VisitingTeam: lgsrcTeam,
		}

		_, first, outcome, locator, second := m.ForOrganization(asrc)
		require.Contains(t, first, "LGSRC")
		require.Equal(t, "", outcome)
		require.Equal(t, "@", locator)
		require.Equal(t, "Almaden Valley Athletic Club", second)
	})
}
