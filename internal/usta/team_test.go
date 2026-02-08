package usta

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseTime(t *testing.T) {
	cases := map[string]struct {
		hour   int
		minute int
		error  string
	}{
		"All 3 at 7:30 PM": {19, 30, ""},
		"All 3 at 9:30 AM backup(Sundays ) if raining":          {9, 30, ""},
		"3/1 at 6:30 PM and 7:45 PM Gate Code 24865":            {18, 30, ""},
		"All 3 at 2:00 PM Courts 7, 8 and 9":                    {14, 0, ""},
		"All 3 at 11:00 AM Warm up court available at 10:30am.": {11, 0, ""},
		"All 3 at 12:00 PM CTS 3,4,5":                           {12, 0, ""},
		"All 4 at 12:30 PM":                                     {12, 30, ""},
		"All 3 at 12:00 AM":                                     {0, 0, ""},
		"All 3 at 12:30 AM":                                     {0, 30, ""},
	}

	for input, test := range cases {
		t.Run(input, func(t *testing.T) {
			hour, min, err := parseTime(input)
			if test.error != "" {
				require.Equal(t, test.error, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, test.hour, hour)
				require.Equal(t, test.minute, min)
			}
		})
	}
}
