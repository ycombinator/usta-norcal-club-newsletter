package formatters

import (
	"fmt"
	"io"
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
		recent := BuildRecentResultsData(data.Org, data.PastMatches, data.OrgNames, f.reader, f.writer)
		html, err := RenderRecentResultsHTML(recent)
		if err != nil {
			return fmt.Errorf("rendering recent results HTML: %w", err)
		}
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
		fmt.Fprintln(f.writer, "Wrote", path)
	}

	if len(data.FutureMatches) > 0 {
		upcoming := BuildUpcomingMatchesData(data.Org, data.FutureMatches, data.OrgNames, f.reader, f.writer)
		html, err := RenderUpcomingMatchesHTML(upcoming)
		if err != nil {
			return fmt.Errorf("rendering upcoming matches HTML: %w", err)
		}
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
		fmt.Fprintln(f.writer, "Wrote", path)
	}

	return data.Save()
}
