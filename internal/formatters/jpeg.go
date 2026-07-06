package formatters

import (
	"fmt"
	"log/slog"
	"os"
)

type JPEGFormatter struct{}

func NewJPEGFormatter() *JPEGFormatter {
	return &JPEGFormatter{}
}

func (f *JPEGFormatter) FormatRecent(data *PreparedData, cfg Config) error {
	if !data.hasPastMatches() {
		return nil
	}

	recent := data.buildRecentDisplay(cfg)
	slog.Info("rendering recent results", "rows", len(recent.Rows))
	html, err := RenderRecentResultsHTML(recent)
	if err != nil {
		return fmt.Errorf("rendering recent results HTML: %w", err)
	}
	slog.Info("capturing recent results screenshot")
	jpeg, err := renderHTMLToJPEG(html, 90)
	if err != nil {
		return fmt.Errorf("rendering recent results JPEG: %w", err)
	}
	path, err := OutputPath(cfg.OutputDir, OutputFilename(data.orgShortName(), "recent", "jpg"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, jpeg, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	slog.Info("wrote recent results", "path", path, "size_bytes", len(jpeg))
	fmt.Fprintln(cfg.Writer, "Wrote", path)

	return nil
}

func (f *JPEGFormatter) FormatUpcoming(data *PreparedData, cfg Config) error {
	if !data.hasUpcomingMatches() {
		return nil
	}

	upcoming := data.buildUpcomingDisplay(cfg)
	slog.Info("rendering upcoming matches", "days", len(upcoming.Days))
	html, err := RenderUpcomingMatchesHTML(upcoming)
	if err != nil {
		return fmt.Errorf("rendering upcoming matches HTML: %w", err)
	}
	slog.Info("capturing upcoming matches screenshot")
	jpeg, err := renderHTMLToJPEG(html, 90)
	if err != nil {
		return fmt.Errorf("rendering upcoming matches JPEG: %w", err)
	}
	path, err := OutputPath(cfg.OutputDir, OutputFilename(data.orgShortName(), "upcoming", "jpg"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, jpeg, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	slog.Info("wrote upcoming matches", "path", path, "size_bytes", len(jpeg))
	fmt.Fprintln(cfg.Writer, "Wrote", path)

	return nil
}
