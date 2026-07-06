package usta

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisplay(t *testing.T) {
	tests := map[string]struct {
		name        string
		code        string
		gender      Gender
		level       string
		teamSuffix  string
		daytime     bool
		genderEmoji string
	}{
		"18+ womens": {
			name:        "2026 Adult 18+ Womens 3.5",
			gender:      GenderWomens,
			level:       "3.5",
			daytime:     false,
			genderEmoji: "👭",
		},
		"18+ mens": {
			name:        "2026 Adult 18+ Mens 4.0",
			gender:      GenderMens,
			level:       "4.0",
			daytime:     false,
			genderEmoji: "👬",
		},
		"18+ mixed": {
			name:        "2026 Adult 18+ Mixed 9.0",
			gender:      GenderMixed,
			level:       "9.0",
			daytime:     false,
			genderEmoji: "👫",
		},
		"55 & over womens": {
			name:        "2026 Adult 55 & Over Womens 3.5",
			gender:      GenderWomens,
			level:       "3.5",
			daytime:     false,
			genderEmoji: "👭",
		},
		"40 & over mens": {
			name:        "2026 Adult 40 & Over Mens 4.0",
			gender:      GenderMens,
			level:       "4.0",
			daytime:     false,
			genderEmoji: "👬",
		},
		"55 & over with ntrp & over": {
			name:        "2026 Adult 55 & Over Womens 2.5 & Over",
			gender:      GenderWomens,
			level:       "2.5+",
			daytime:     false,
			genderEmoji: "👭",
		},
		"65 & over": {
			name:        "2026 Adult 65 & Over Mens 3.0",
			gender:      GenderMens,
			level:       "3.0",
			daytime:     false,
			genderEmoji: "👬",
		},
		"18+ daytime": {
			name:        "2026 Adult 18+ Womens 3.5 Daytime",
			gender:      GenderWomens,
			level:       "3.5",
			daytime:     true,
			genderEmoji: "👭",
		},
		"18+ daytime with ntrp plus": {
			name:        "2026 Adult 18+ Womens 2.5+ Daytime",
			gender:      GenderWomens,
			level:       "2.5+",
			daytime:     true,
			genderEmoji: "👭",
		},
		"no year prefix": {
			name:        "Adult 18+ Womens 3.5",
			gender:      GenderWomens,
			level:       "3.5",
			daytime:     false,
			genderEmoji: "👭",
		},
		"mixed 40 & over (gender first)": {
			name:        "2026 Mixed 40 & Over 7.0",
			gender:      GenderMixed,
			level:       "7.0",
			daytime:     false,
			genderEmoji: "👫",
		},
		"unknown format": {
			name:        "Some Random Team Name",
			gender:      GenderUnknown,
			level:       "",
			daytime:     false,
			genderEmoji: "",
		},
		"mixed 40 & over team A (gender first)": {
			name:        "2026 Mixed 40 & Over 7.0",
			code:        "CLUB SR 40MX7.0A",
			gender:      GenderMixed,
			level:       "7.0",
			teamSuffix:  "A",
			daytime:     false,
			genderEmoji: "👫",
		},
		"mixed 40 & over team B (gender first)": {
			name:        "2026 Mixed 40 & Over 7.0",
			code:        "CLUB SR 40MX7.0B",
			gender:      GenderMixed,
			level:       "7.0",
			teamSuffix:  "B",
			daytime:     false,
			genderEmoji: "👫",
		},
		"18+ mixed team C": {
			name:        "2026 Adult 18+ Mixed 7.0",
			code:        "CLUB SR 18MX7.0C",
			gender:      GenderMixed,
			level:       "7.0",
			teamSuffix:  "C",
			daytime:     false,
			genderEmoji: "👫",
		},
		"daytime team with suffix": {
			name:        "2026 Adult 18+ Womens 3.5 Daytime",
			code:        "CLUB SR 18AW3.5A-DT",
			gender:      GenderWomens,
			level:       "3.5",
			teamSuffix:  "A",
			daytime:     true,
			genderEmoji: "👭",
		},
		"team B with bracketed nickname": {
			name:        "2026 Mixed 40 & Over 7.0",
			code:        "CLUB SR 40MX7.0B [Team Bee's]",
			gender:      GenderMixed,
			level:       "7.0",
			teamSuffix:  "B",
			daytime:     false,
			genderEmoji: "👫",
		},
		"team A with parenthesized nickname": {
			name:        "2026 Mixed 40 & Over 6.0",
			code:        "CLUB SR 40MX6.0A (Summer Swingers)",
			gender:      GenderMixed,
			level:       "6.0",
			teamSuffix:  "A",
			daytime:     false,
			genderEmoji: "👫",
		},
	}

	for label, tc := range tests {
		t.Run(label, func(t *testing.T) {
			team := &Team{Name: tc.name, Code: tc.code}
			d := team.Display()
			require.Equal(t, tc.gender, d.Gender)
			require.Equal(t, tc.level, d.Level)
			require.Equal(t, tc.teamSuffix, d.TeamSuffix)
			require.Equal(t, tc.daytime, d.Daytime)
			require.Equal(t, tc.genderEmoji, d.GenderEmoji())
		})
	}
}
