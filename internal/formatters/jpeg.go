package formatters

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
)

type JPEGFormatter struct {
	reader io.Reader
	writer io.Writer
}

func NewJPEGFormatter() *JPEGFormatter {
	return &JPEGFormatter{reader: os.Stdin, writer: os.Stdout}
}

func (f *JPEGFormatter) Format(n *core.Newsletter, cfg Config) error {
	data, err := Prepare(n, cfg, f.reader, f.writer)
	if err != nil {
		return err
	}

	orgName := data.Org.ShortName()

	if len(data.PastMatches) > 0 {
		slog.Info("rendering recent results", "matches", len(data.PastMatches))
		recent := BuildRecentResultsData(data.Org, data.PastMatches, data.OrgNames, f.reader, f.writer)
		html, err := RenderRecentResultsHTML(recent)
		if err != nil {
			return fmt.Errorf("rendering recent results HTML: %w", err)
		}
		slog.Info("capturing recent results screenshot")
		jpeg, err := renderHTMLToJPEG(html, 90)
		if err != nil {
			return fmt.Errorf("rendering recent results JPEG: %w", err)
		}
		path, err := OutputPath(cfg.OutputDir, OutputFilename(orgName, "recent", "jpg"))
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, jpeg, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		slog.Info("wrote recent results", "path", path, "size_bytes", len(jpeg))
		fmt.Fprintln(f.writer, "Wrote", path)
	}

	if len(data.FutureMatches) > 0 {
		slog.Info("rendering upcoming matches", "matches", len(data.FutureMatches))
		upcoming := BuildUpcomingMatchesData(data.Org, data.FutureMatches, data.OrgNames, data.LocationOverrides, f.reader, f.writer)
		html, err := RenderUpcomingMatchesHTML(upcoming)
		if err != nil {
			return fmt.Errorf("rendering upcoming matches HTML: %w", err)
		}
		slog.Info("capturing upcoming matches screenshot")
		jpeg, err := renderHTMLToJPEG(html, 90)
		if err != nil {
			return fmt.Errorf("rendering upcoming matches JPEG: %w", err)
		}
		path, err := OutputPath(cfg.OutputDir, OutputFilename(orgName, "upcoming", "jpg"))
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, jpeg, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		slog.Info("wrote upcoming matches", "path", path, "size_bytes", len(jpeg))
		fmt.Fprintln(f.writer, "Wrote", path)
	}

	return data.Save()
}
