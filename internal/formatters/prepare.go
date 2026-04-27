package formatters

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

type PreparedData struct {
	Org           *usta.Organization
	PastMatches   []AnnotatedMatch
	FutureMatches []usta.Match
	OrgNames      *OrgNames
}

func Prepare(n *core.Newsletter, cfg Config, reader io.Reader, writer io.Writer) (*PreparedData, error) {
	org := n.Organization()
	pastMatches, futureMatches := org.Matches(cfg.PastDuration, cfg.FutureDuration)
	slog.Info("filtered matches", "past", len(pastMatches), "future", len(futureMatches))

	annotated := make([]AnnotatedMatch, len(pastMatches))
	for i, m := range pastMatches {
		annotated[i] = AnnotatedMatch{Match: m}
	}

	slog.Info("loading org display names", "file", orgNamesFile)
	names, err := LoadOrgNames()
	if err != nil {
		return nil, fmt.Errorf("loading org names: %w", err)
	}
	slog.Info("loaded org display names", "count", len(names.names))

	PromptNoOutcomeMatches(reader, writer, annotated, org, names)
	PromptPlayoffMatches(reader, writer, annotated, org, names)

	return &PreparedData{
		Org:           org,
		PastMatches:   annotated,
		FutureMatches: futureMatches,
		OrgNames:      names,
	}, nil
}

func (d *PreparedData) Save() error {
	if err := d.OrgNames.Save(); err != nil {
		return fmt.Errorf("saving org names: %w", err)
	}
	return nil
}

func OutputFilename(orgShortName, suffix, ext string) string {
	now := time.Now()
	return fmt.Sprintf("%s_usta_%s_%s.%s",
		strings.ToLower(orgShortName),
		now.Format("2006_01_02"),
		suffix,
		ext,
	)
}

func OutputPath(dir, filename string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating output directory %s: %w", dir, err)
	}
	return filepath.Join(dir, filename), nil
}
