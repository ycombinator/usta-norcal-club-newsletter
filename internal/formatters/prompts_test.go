package formatters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

func makeTestOrg() *usta.Organization {
	return &usta.Organization{ID: 225, Name: "Almaden Swim Racquet Club"}
}

func makeTestOrgNames() *OrgNames {
	return &OrgNames{names: map[string]string{
		"ALMADEN VALLEY ATHLETIC CLUB": "AVAC",
	}}
}

func makeNoOutcomeMatch() AnnotatedMatch {
	org := makeTestOrg()
	opponent := &usta.Organization{ID: 300, Name: "Almaden Valley Athletic Club"}
	return AnnotatedMatch{
		Match: usta.Match{
			HomeTeam:     &usta.Team{ID: 1, Name: "Adult 18+ Womens 3.5", Organization: org},
			VisitingTeam: &usta.Team{ID: 2, Name: "Adult 18+ Mens 4.0", Organization: opponent},
		},
	}
}

func makeWinMatch() AnnotatedMatch {
	org := makeTestOrg()
	opponent := &usta.Organization{ID: 300, Name: "Almaden Valley Athletic Club"}
	homeTeam := &usta.Team{ID: 1, Name: "Adult 18+ Womens 3.5", Organization: org}
	m := AnnotatedMatch{
		Match: usta.Match{
			HomeTeam:     homeTeam,
			VisitingTeam: &usta.Team{ID: 2, Name: "Adult 18+ Mens 4.0", Organization: opponent},
		},
	}
	m.Match.Outcome.WinningTeam = homeTeam
	m.Match.Outcome.WinnerPoints = 3
	m.Match.Outcome.LoserPoints = 2
	return m
}

func TestPromptNoOutcomeMatches_RainedOut(t *testing.T) {
	matches := []AnnotatedMatch{makeNoOutcomeMatch()}
	input := strings.NewReader("R\n")
	output := &bytes.Buffer{}

	PromptNoOutcomeMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.True(t, matches[0].Annotation.RainedOut)
	require.Empty(t, matches[0].Annotation.Score)
	require.Empty(t, matches[0].Annotation.Footnote)
}

func TestPromptNoOutcomeMatches_RainedOutLowercase(t *testing.T) {
	matches := []AnnotatedMatch{makeNoOutcomeMatch()}
	input := strings.NewReader("r\n")
	output := &bytes.Buffer{}

	PromptNoOutcomeMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.True(t, matches[0].Annotation.RainedOut)
}

func TestPromptNoOutcomeMatches_ScoreWithExplanation(t *testing.T) {
	matches := []AnnotatedMatch{makeNoOutcomeMatch()}
	input := strings.NewReader("2-1 to be completed at a later date\n")
	output := &bytes.Buffer{}

	PromptNoOutcomeMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.False(t, matches[0].Annotation.RainedOut)
	require.Equal(t, "2-1", matches[0].Annotation.Score)
	require.Equal(t, "to be completed at a later date", matches[0].Annotation.Footnote)
}

func TestPromptNoOutcomeMatches_ExplanationOnly(t *testing.T) {
	matches := []AnnotatedMatch{makeNoOutcomeMatch()}
	input := strings.NewReader("match cancelled\n")
	output := &bytes.Buffer{}

	PromptNoOutcomeMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.False(t, matches[0].Annotation.RainedOut)
	require.Empty(t, matches[0].Annotation.Score)
	require.Equal(t, "match cancelled", matches[0].Annotation.Footnote)
}

func TestPromptNoOutcomeMatches_InvalidThenValid(t *testing.T) {
	matches := []AnnotatedMatch{makeNoOutcomeMatch()}
	input := strings.NewReader("\n2-1\nR\n")
	output := &bytes.Buffer{}

	PromptNoOutcomeMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.True(t, matches[0].Annotation.RainedOut)
	require.Contains(t, output.String(), "Invalid input")
}

func TestPromptNoOutcomeMatches_SkipsMatchesWithOutcome(t *testing.T) {
	matches := []AnnotatedMatch{makeWinMatch()}
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	PromptNoOutcomeMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.False(t, matches[0].Annotation.RainedOut)
	require.Empty(t, output.String())
}

func TestPromptPlayoffMatches_No(t *testing.T) {
	matches := []AnnotatedMatch{makeWinMatch()}
	input := strings.NewReader("n\n")
	output := &bytes.Buffer{}

	PromptPlayoffMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.Equal(t, RegularSeason, matches[0].Annotation.MatchType)
}

func TestPromptPlayoffMatches_YesThenClassify(t *testing.T) {
	matches := []AnnotatedMatch{makeWinMatch(), makeWinMatch(), makeWinMatch()}
	input := strings.NewReader("y\np\ns\nr\n")
	output := &bytes.Buffer{}

	PromptPlayoffMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.Equal(t, Playoff, matches[0].Annotation.MatchType)
	require.Equal(t, Sectionals, matches[1].Annotation.MatchType)
	require.Equal(t, RegularSeason, matches[2].Annotation.MatchType)
}

func TestPromptPlayoffMatches_InvalidThenValid(t *testing.T) {
	matches := []AnnotatedMatch{makeWinMatch()}
	input := strings.NewReader("y\nx\np\n")
	output := &bytes.Buffer{}

	PromptPlayoffMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.Equal(t, Playoff, matches[0].Annotation.MatchType)
	require.Contains(t, output.String(), "Please enter p, s, or r")
}

func TestPromptPlayoffMatches_Empty(t *testing.T) {
	var matches []AnnotatedMatch
	input := strings.NewReader("")
	output := &bytes.Buffer{}

	PromptPlayoffMatches(input, output, matches, makeTestOrg(), makeTestOrgNames())

	require.Empty(t, output.String())
}
