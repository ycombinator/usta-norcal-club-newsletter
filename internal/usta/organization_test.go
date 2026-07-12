package usta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOrganizationMatchesIncludesPastMatchesWithoutTime(t *testing.T) {
	boundary := time.Date(2026, 7, 12, 0, 0, 0, 0, tz)

	teamA := &Team{ID: 1}
	teamB := &Team{ID: 2}

	pastMatchNoTime := Match{
		Date:         time.Date(2026, 7, 10, 0, 0, 0, 0, tz),
		HasTime:      false,
		HomeTeam:     teamA,
		VisitingTeam: teamB,
	}
	futureMatchNoTime := Match{
		Date:         time.Date(2026, 7, 13, 0, 0, 0, 0, tz),
		HasTime:      false,
		HomeTeam:     teamA,
		VisitingTeam: teamB,
	}
	futureMatchWithTime := Match{
		Date:         time.Date(2026, 7, 13, 18, 30, 0, 0, tz),
		HasTime:      true,
		HomeTeam:     teamA,
		VisitingTeam: teamB,
	}

	teamA.Matches = []Match{pastMatchNoTime, futureMatchNoTime, futureMatchWithTime}

	org := &Organization{Teams: []*Team{teamA, teamB}}

	past, future := org.Matches(7*24*time.Hour, 7*24*time.Hour, boundary)

	require.Len(t, past, 1)
	require.Equal(t, pastMatchNoTime.Date, past[0].Date)

	require.Len(t, future, 1)
	require.True(t, future[0].HasTime)
}
