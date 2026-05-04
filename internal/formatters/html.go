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
	if len(data.PastMatches) == 0 {
		return nil
	}

	orgName := data.Org.ShortName()

	recent := BuildRecentResultsData(data.Org, data.PastMatches, data.OrgNames, cfg.Reader, cfg.Writer)
	html, err := RenderRecentResultsHTML(recent)
	if err != nil {
		return fmt.Errorf("rendering recent results HTML: %w", err)
	}
	path, err := OutputPath(cfg.OutputDir, OutputFilename(orgName, "recent", "html"))
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
	if len(data.FutureMatches) == 0 {
		return nil
	}

	orgName := data.Org.ShortName()

	upcoming := BuildUpcomingMatchesData(data.Org, data.FutureMatches, data.OrgNames, data.LocationOverrides, cfg.Reader, cfg.Writer)
	html, err := RenderUpcomingMatchesHTML(upcoming)
	if err != nil {
		return fmt.Errorf("rendering upcoming matches HTML: %w", err)
	}
	path, err := OutputPath(cfg.OutputDir, OutputFilename(orgName, "upcoming", "html"))
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(html), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Fprintln(cfg.Writer, "Wrote", path)

	return nil
}
