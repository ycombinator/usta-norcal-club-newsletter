package usta

import (
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

var teamNameRegex = regexp.MustCompile(
	`(?i)Adult\s+(\d+)(?:\s*\+|\s+&\s+Over)\s+(Womens|Women'?s|Mens|Men'?s|Mixed)\s+(\d+\.?\d*)(\s*\+|\s+&\s+Over)?`,
)

var genderFirstRegex = regexp.MustCompile(
	`(?i)(Womens|Women'?s|Mens|Men'?s|Mixed)\s+\d+(?:\s*\+|\s+&\s+Over)\s+(\d+\.?\d*)(\s*\+|\s+&\s+Over)?`,
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
	} else {
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
