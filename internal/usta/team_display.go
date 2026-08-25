package usta

import (
	"log/slog"
	"regexp"
	"strings"
)

type Gender int

const (
	GenderUnknown Gender = iota
	GenderWomens
	GenderMens
	GenderMixed
)

type TeamDisplay struct {
	Gender     Gender
	Level      string
	TeamSuffix string
	Daytime    bool
}

// genderWords matches the gender component of a USTA team name. USTA NorCal
// uses "Mens"/"Womens" (most common), "Men's"/"Women's" (ASCII or smart
// apostrophe \x{2019}), and occasionally the singular "Men"/"Women".
const genderWords = `(Womens|Women[\x{2019}']?s?|Mens|Men[\x{2019}']?s?|Mixed)`

var teamNameRegex = regexp.MustCompile(
	`(?i)Adult\s+(\d+)(?:\s*\+|\s+&\s+Over)\s+` + genderWords + `\s+(\d+\.?\d*)(\s*\+|\s+&\s+Over)?`,
)

var genderFirstRegex = regexp.MustCompile(
	`(?i)` + genderWords + `\s+\d+(?:\s*\+|\s+&\s+Over)\s+(\d+\.?\d*)(\s*\+|\s+&\s+Over)?`,
)

// comboRegex matches Combo Doubles league names: "2026 Combo Mens League 6.5"
// or "2026 Combo Womens Daytime League 5.5". There is no age category.
var comboRegex = regexp.MustCompile(
	`(?i)Combo\s+` + genderWords + `\s+(?:Daytime\s+)?League\s+(\d+\.?\d*)`,
)

// teamCodeSuffixRegex extracts the trailing team letter (A, B, C…) from a short
// code like "ALMADEN SR 40MX7.0A" or "ALMADEN SR 18AW2.5+A-DT".
var teamCodeSuffixRegex = regexp.MustCompile(`\d+\.?\d*\+?([A-Z])(?:-DT)?$`)

func extractTeamSuffix(code string) string {
	if idx := strings.IndexAny(code, "(["); idx >= 0 {
		code = strings.TrimSpace(code[:idx])
	}
	m := teamCodeSuffixRegex.FindStringSubmatch(code)
	if m == nil {
		return ""
	}
	return m[1]
}

func (t *Team) Display() TeamDisplay {
	var genderStr, levelStr, suffixStr string

	if m := teamNameRegex.FindStringSubmatch(t.Name); m != nil {
		genderStr, levelStr, suffixStr = m[2], m[3], m[4]
	} else if m := genderFirstRegex.FindStringSubmatch(t.Name); m != nil {
		genderStr, levelStr, suffixStr = m[1], m[2], m[3]
	} else if m := comboRegex.FindStringSubmatch(t.Name); m != nil {
		genderStr, levelStr = m[1], m[2]
	} else {
		if t.Name != "" {
			slog.Warn("team name did not match any known format", "team_id", t.ID, "team_name", t.Name)
		}
		return TeamDisplay{}
	}

	var gender Gender
	switch {
	case len(genderStr) > 0 && (genderStr[0] == 'W' || genderStr[0] == 'w'):
		gender = GenderWomens
	case len(genderStr) > 0 && (genderStr[0] == 'M' && (genderStr[1] == 'e' || genderStr[1] == 'E')):
		gender = GenderMens
	default:
		gender = GenderMixed
	}

	level := levelStr
	if suffixStr != "" {
		level += "+"
	}

	return TeamDisplay{
		Gender:     gender,
		Level:      level,
		TeamSuffix: extractTeamSuffix(t.Code),
		Daytime:    strings.Contains(strings.ToLower(t.Name), "daytime"),
	}
}

func (d TeamDisplay) GenderEmoji() string {
	switch d.Gender {
	case GenderWomens:
		return "👭"
	case GenderMens:
		return "👬"
	case GenderMixed:
		return "👫"
	default:
		return ""
	}
}

func (d TeamDisplay) DaytimeEmoji() string {
	if d.Daytime {
		return "☀️"
	}
	return ""
}
