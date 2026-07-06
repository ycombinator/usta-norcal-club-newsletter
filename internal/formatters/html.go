package formatters

import (
	"fmt"
	"os"
)

type HTMLFormatter struct{}

func NewHTMLFormatter() *HTMLFormatter {
	return &HTMLFormatter{}
}

func (f *HTMLFormatter) FormatRecent(data *PreparedData, cfg Config) error {
	if !data.hasPastMatches() {
		return nil
	}

	recent := data.buildRecentDisplay(cfg)
	html, err := RenderRecentResultsHTML(recent)
	if err != nil {
		return fmt.Errorf("rendering recent results HTML: %w", err)
	}
	path, err := OutputPath(cfg.OutputDir, OutputFilename(data.orgShortName(), "recent", "html"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(html), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Fprintln(cfg.Writer, "Wrote", path)

	return nil
}

func (f *HTMLFormatter) FormatUpcoming(data *PreparedData, cfg Config) error {
	if !data.hasUpcomingMatches() {
		return nil
	}

	upcoming := data.buildUpcomingDisplay(cfg)
	html, err := RenderUpcomingMatchesHTML(upcoming)
	if err != nil {
		return fmt.Errorf("rendering upcoming matches HTML: %w", err)
	}
	path, err := OutputPath(cfg.OutputDir, OutputFilename(data.orgShortName(), "upcoming", "html"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(html), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Fprintln(cfg.Writer, "Wrote", path)

	return nil
}
