package formatters

import "github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"

type MatchType int

const (
	RegularSeason MatchType = iota
	Playoff
	Sectionals
)

type MatchAnnotation struct {
	RainedOut  bool
	Score      string
	Footnote   string
	MatchType  MatchType
}

type AnnotatedMatch struct {
	Match      usta.Match
	Annotation MatchAnnotation
}
