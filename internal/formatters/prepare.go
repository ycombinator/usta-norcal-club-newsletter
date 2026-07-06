package formatters

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

type PreparedData struct {
	// Populated when fetching live from USTA (nil when loaded from data file).
	Org               *usta.Organization
	PastMatches       []AnnotatedMatch
	FutureMatches     []usta.Match
	OrgNames          *OrgNames
	LocationOverrides map[int]string

	// Non-nil when loaded from an existing data file instead of USTA.
	DataFile *DataFile
}

// orgShortName returns the org short name from either live data or the data file.
func (d *PreparedData) orgShortName() string {
	if d.DataFile != nil {
		return d.DataFile.OrgShortName
	}
	return d.Org.ShortName()
}

// hasPastMatches reports whether there are any past matches to render.
func (d *PreparedData) hasPastMatches() bool {
	if d.DataFile != nil {
		return len(d.DataFile.PastMatches) > 0
	}
	return len(d.PastMatches) > 0
}

// hasUpcomingMatches reports whether there are any upcoming matches to render.
func (d *PreparedData) hasUpcomingMatches() bool {
	if d.DataFile != nil {
		return len(d.DataFile.FutureMatches) > 0
	}
	return len(d.FutureMatches) > 0
}

// buildRecentDisplay returns display-ready recent results data.
// Uses the data file when available, otherwise builds from live match data.
func (d *PreparedData) buildRecentDisplay(cfg Config) RecentResultsData {
	if d.DataFile != nil {
		return d.DataFile.ToRecentResultsData()
	}
	return BuildRecentResultsData(d.Org, d.PastMatches, d.OrgNames, cfg.Reader, cfg.Writer)
}

// buildUpcomingDisplay returns display-ready upcoming matches data.
// Uses the data file when available, otherwise builds from live match data.
func (d *PreparedData) buildUpcomingDisplay(cfg Config) UpcomingMatchesData {
	if d.DataFile != nil {
		return d.DataFile.ToUpcomingMatchesData()
	}
	return BuildUpcomingMatchesData(d.Org, d.FutureMatches, d.OrgNames, d.LocationOverrides, cfg.Reader, cfg.Writer)
}

func Prepare(n *core.Newsletter, cfg Config) (*PreparedData, error) {
	// If a data file already exists, load it and skip fetching from USTA.
	if cfg.DataFilePath != "" {
		if _, err := os.Stat(cfg.DataFilePath); err == nil {
			slog.Info("data file found, loading instead of fetching from USTA", "path", cfg.DataFilePath)
			df, err := LoadDataFile(cfg.DataFilePath)
			if err != nil {
				return nil, fmt.Errorf("loading data file: %w", err)
			}
			fmt.Fprintln(cfg.Writer, "Loaded data file", cfg.DataFilePath)
			return &PreparedData{DataFile: df}, nil
		}
	}

	org := n.Organization()
	pastMatches, futureMatches := org.Matches(cfg.PastDuration, cfg.FutureDuration, cfg.BoundaryDate)
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

	PromptNoOutcomeMatches(cfg.Reader, cfg.Writer, annotated, org, names)
	PromptPlayoffMatches(cfg.Reader, cfg.Writer, annotated, org, names)
	locationOverrides := PromptExtraTeamLocations(cfg.Reader, cfg.Writer, futureMatches, org, names)

	data := &PreparedData{
		Org:               org,
		PastMatches:       annotated,
		FutureMatches:     futureMatches,
		OrgNames:          names,
		LocationOverrides: locationOverrides,
	}

	// Save intermediate data file so it can be edited and re-used.
	if cfg.DataFilePath != "" {
		df := NewDataFile(data, cfg.Reader, cfg.Writer)
		if err := df.Save(cfg.DataFilePath); err != nil {
			slog.Warn("failed to save data file", "path", cfg.DataFilePath, "error", err)
		} else {
			slog.Info("saved data file", "path", cfg.DataFilePath)
			fmt.Fprintln(cfg.Writer, "Wrote", cfg.DataFilePath)
		}
	}

	return data, nil
}

func (d *PreparedData) Save() error {
	if d.OrgNames == nil {
		return nil
	}
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
