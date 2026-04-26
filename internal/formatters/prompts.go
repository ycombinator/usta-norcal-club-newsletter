package formatters

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

var scoreExplanationRegex = regexp.MustCompile(`^(\d+-\d+)\s+(.+)$`)

func describeMatch(m usta.Match, org *usta.Organization, names *OrgNames, reader io.Reader, writer io.Writer) string {
	ourTeam, opponent, isHome := resolveTeams(m, org)
	opponent.LoadOrganization(context.Background())

	d := ourTeam.Display()
	opName := opponentDisplayName(names, reader, writer, opponent.Organization)

	locator := "vs"
	if !isHome {
		locator = "@"
	}

	return fmt.Sprintf("%s: %s%s %s %s",
		m.Date.Format("Mon 1/2"),
		d.GenderEmoji(), d.Level,
		locator,
		opName,
	)
}

func PromptNoOutcomeMatches(reader io.Reader, writer io.Writer, matches []AnnotatedMatch, org *usta.Organization, names *OrgNames) {
	scanner := bufio.NewScanner(reader)

	for i := range matches {
		if matches[i].Match.Outcome.WinningTeam != nil {
			continue
		}

		desc := describeMatch(matches[i].Match, org, names, reader, writer)

		for {
			fmt.Fprintf(writer, "%s — no outcome recorded.\n", desc)
			fmt.Fprintf(writer, "Enter R if rained out, or score and explanation (e.g. \"2-1 to be completed later\"): ")

			if !scanner.Scan() {
				return
			}
			input := strings.TrimSpace(scanner.Text())

			if strings.EqualFold(input, "r") {
				matches[i].Annotation.RainedOut = true
				break
			}

			if m := scoreExplanationRegex.FindStringSubmatch(input); m != nil {
				matches[i].Annotation.Score = m[1]
				matches[i].Annotation.Footnote = m[2]
				break
			}

			if len(input) > 0 && (input[0] < '0' || input[0] > '9') {
				matches[i].Annotation.Footnote = input
				break
			}

			fmt.Fprintf(writer, "Invalid input. Enter R, or \"score explanation\", or an explanation.\n")
		}
	}
}

func PromptPlayoffMatches(reader io.Reader, writer io.Writer, matches []AnnotatedMatch, org *usta.Organization, names *OrgNames) {
	if len(matches) == 0 {
		return
	}

	scanner := bufio.NewScanner(reader)

	for {
		fmt.Fprintf(writer, "Were any recent matches playoffs or Sectionals? (y/n): ")
		if !scanner.Scan() {
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if strings.EqualFold(input, "n") {
			return
		}
		if strings.EqualFold(input, "y") {
			break
		}
		fmt.Fprintf(writer, "Please enter y or n.\n")
	}

	for i := range matches {
		desc := describeMatch(matches[i].Match, org, names, reader, writer)

		for {
			fmt.Fprintf(writer, "%s — (p)layoff / (s)ectionals / (r)egular: ", desc)
			if !scanner.Scan() {
				return
			}
			input := strings.TrimSpace(scanner.Text())

			switch strings.ToLower(input) {
			case "p":
				matches[i].Annotation.MatchType = Playoff
			case "s":
				matches[i].Annotation.MatchType = Sectionals
			case "r":
				matches[i].Annotation.MatchType = RegularSeason
			default:
				fmt.Fprintf(writer, "Please enter p, s, or r.\n")
				continue
			}
			break
		}
	}
}
