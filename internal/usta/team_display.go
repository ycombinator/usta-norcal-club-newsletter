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
	Gender  Gender
	Level   string
	Daytime bool
}

var teamNameRegex = regexp.MustCompile(
	`(?i)Adult\s+(\d+)(?:\s*\+|\s+&\s+Over)\s+(Womens|Women'?s|Mens|Men'?s|Mixed)\s+(\d+\.?\d*)(\s*\+|\s+&\s+Over)?`,
)

func (t *Team) Display() TeamDisplay {
	matches := teamNameRegex.FindStringSubmatch(t.Name)
	if matches == nil {
		return TeamDisplay{}
	}

	var gender Gender
	switch {
	case len(matches[2]) > 0 && (matches[2][0] == 'W' || matches[2][0] == 'w'):
		gender = GenderWomens
	case len(matches[2]) > 0 && (matches[2][0] == 'M' && (matches[2][1] == 'e' || matches[2][1] == 'E')):
		gender = GenderMens
	default:
		gender = GenderMixed
	}

	level := matches[3]
	if matches[4] != "" {
		level += "+"
	}

	return TeamDisplay{
		Gender:  gender,
		Level:   level,
		Daytime: strings.Contains(strings.ToLower(t.Name), "daytime"),
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
