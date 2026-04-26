package usta

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisplay(t *testing.T) {
	tests := map[string]struct {
		name        string
		gender      Gender
		level       string
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
			daytime:     true,
			genderEmoji: "👭",
		},
		"40 & over mens": {
			name:        "2026 Adult 40 & Over Mens 4.0",
			gender:      GenderMens,
			level:       "4.0",
			daytime:     true,
			genderEmoji: "👬",
		},
		"55 & over with ntrp & over": {
			name:        "2026 Adult 55 & Over Womens 2.5 & Over",
			gender:      GenderWomens,
			level:       "2.5+",
			daytime:     true,
			genderEmoji: "👭",
		},
		"65 & over": {
			name:        "2026 Adult 65 & Over Mens 3.0",
			gender:      GenderMens,
			level:       "3.0",
			daytime:     true,
			genderEmoji: "👬",
		},
		"no year prefix": {
			name:        "Adult 18+ Womens 3.5",
			gender:      GenderWomens,
			level:       "3.5",
			daytime:     false,
			genderEmoji: "👭",
		},
		"unknown format": {
			name:        "Some Random Team Name",
			gender:      GenderUnknown,
			level:       "",
			daytime:     false,
			genderEmoji: "",
		},
	}

	for label, tc := range tests {
		t.Run(label, func(t *testing.T) {
			team := &Team{Name: tc.name}
			d := team.Display()
			require.Equal(t, tc.gender, d.Gender)
			require.Equal(t, tc.level, d.Level)
			require.Equal(t, tc.daytime, d.Daytime)
			require.Equal(t, tc.genderEmoji, d.GenderEmoji())
		})
	}
}
