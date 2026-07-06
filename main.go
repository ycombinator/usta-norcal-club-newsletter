package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/formatters"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: usta-norcal-club-newsletter [flags]

Generate a newsletter of recent and upcoming USTA NorCal tennis matches
for a club organization.

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
  usta-norcal-club-newsletter                                        Use defaults (ASRC, jpeg)
  usta-norcal-club-newsletter -org=300                               Specify a different organization
  usta-norcal-club-newsletter -teams=123,456                         Track additional teams by ID
  usta-norcal-club-newsletter -format=console                        Console output for both sections
  usta-norcal-club-newsletter -recent-format=jpeg -upcoming-format=console
  usta-norcal-club-newsletter -upcoming-format=gcal -gcal-credentials=creds.json -gcal-calendar="USTA Tennis"
  usta-norcal-club-newsletter -past=7 -future=14                     Show 7 days back and 14 days ahead
  usta-norcal-club-newsletter -outdir=./output                       Write files to ./output
  usta-norcal-club-newsletter help                                   Show this help message
`)
}

func makeRecentFormatter(name string) (formatters.RecentFormatter, error) {
	switch name {
	case "console":
		return formatters.NewConsoleFormatter(), nil
	case "pdf":
		return formatters.NewPDFFormatter(), nil
	case "jpeg":
		return formatters.NewJPEGFormatter(), nil
	case "html":
		return formatters.NewHTMLFormatter(), nil
	default:
		return nil, fmt.Errorf("unknown recent format: %s (use 'console', 'pdf', 'jpeg', or 'html')", name)
	}
}

func makeUpcomingFormatter(name, gcalCredentials, gcalCalendar string) (formatters.UpcomingFormatter, error) {
	switch name {
	case "console":
		return formatters.NewConsoleFormatter(), nil
	case "pdf":
		return formatters.NewPDFFormatter(), nil
	case "jpeg":
		return formatters.NewJPEGFormatter(), nil
	case "html":
		return formatters.NewHTMLFormatter(), nil
	case "gcal":
		if gcalCredentials == "" {
			return nil, fmt.Errorf("-gcal-credentials is required when upcoming format is 'gcal'")
		}
		if gcalCalendar == "" {
			return nil, fmt.Errorf("-gcal-calendar is required when upcoming format is 'gcal'")
		}
		return &formatters.GCalFormatter{
			CredentialsFile: gcalCredentials,
			CalendarName:    gcalCalendar,
		}, nil
	default:
		return nil, fmt.Errorf("unknown upcoming format: %s (use 'console', 'pdf', 'jpeg', 'html', or 'gcal')", name)
	}
}

func main() {
	c := internal.DefaultConfig()

	flag.Usage = usage
	orgID := flag.Int("org", c.OrganizationID, "USTA NorCal organization ID")
	teams := flag.String("teams", "", "comma-separated list of additional team IDs to track")
	format := flag.String("format", "jpeg", "output format for both sections: console, pdf, jpeg, or html")
	recentFormat := flag.String("recent-format", "", "output format for recent results (overrides -format)")
	upcomingFormat := flag.String("upcoming-format", "", "output format for upcoming matches (overrides -format)")
	pastDays := flag.Int("past", int(c.PastDuration.Hours()/24), "number of days back to include past match results")
	futureDays := flag.Int("future", int(c.FutureDuration.Hours()/24), "number of days ahead to include upcoming matches")
	outDir := flag.String("outdir", "", "output directory for file-based formatters")
	boundaryDate := flag.String("boundary-date", "", "date (YYYY-MM-DD) dividing recent and upcoming matches (default: tomorrow)")
	gcalCredentials := flag.String("gcal-credentials", "", "path to Google OAuth2 client credentials JSON (required for gcal format)")
	gcalCalendar := flag.String("gcal-calendar", "", "Google Calendar name for upcoming match events (required for gcal format)")

	// Handle "help" sub-command before flag.Parse
	if len(os.Args) > 1 && os.Args[1] == "help" {
		usage()
		return
	}

	flag.Parse()

	c.OrganizationID = *orgID
	c.PastDuration = time.Duration(*pastDays) * 24 * time.Hour
	c.FutureDuration = time.Duration(*futureDays) * 24 * time.Hour

	var parsedBoundary time.Time
	if *boundaryDate != "" {
		var err error
		parsedBoundary, err = time.ParseInLocation("2006-01-02", *boundaryDate, time.Now().Location())
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid boundary-date %q: expected YYYY-MM-DD\n", *boundaryDate)
			os.Exit(1)
		}
	}

	if *outDir == "" {
		dirDate := time.Now()
		if !parsedBoundary.IsZero() {
			dirDate = parsedBoundary
		}
		*outDir = filepath.Join(
			os.Getenv("HOME"), "Documents", "ASRC",
			dirDate.Format("2006"),
			dirDate.Format("20060102"),
		)
	}

	if *teams != "" {
		for _, s := range strings.Split(*teams, ",") {
			id, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid team ID %q: %v\n", s, err)
				os.Exit(1)
			}
			c.TeamIDs = append(c.TeamIDs, id)
		}
	}

	if *format == "gcal" {
		fmt.Fprintln(os.Stderr, "'gcal' format is only valid for -upcoming-format, not -format")
		os.Exit(1)
	}
	if *recentFormat == "gcal" {
		fmt.Fprintln(os.Stderr, "'gcal' format is only valid for -upcoming-format, not -recent-format")
		os.Exit(1)
	}

	effectiveRecent := *format
	effectiveUpcoming := *format
	if *recentFormat != "" {
		effectiveRecent = *recentFormat
	}
	if *upcomingFormat != "" {
		effectiveUpcoming = *upcomingFormat
	}

	var err error
	c.RecentFormatter, err = makeRecentFormatter(effectiveRecent)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	c.UpcomingFormatter, err = makeUpcomingFormatter(effectiveUpcoming, *gcalCredentials, *gcalCalendar)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	slog.Info("starting newsletter generation",
		"org", c.OrganizationID,
		"extra_teams", c.TeamIDs,
		"recent_format", effectiveRecent,
		"upcoming_format", effectiveUpcoming,
		"past_days", *pastDays,
		"future_days", *futureDays,
		"outdir", *outDir,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dataFilePath := filepath.Join(*outDir, "data.json")

	fmtCfg := formatters.Config{
		OrganizationID: c.OrganizationID,
		PastDuration:   c.PastDuration,
		FutureDuration: c.FutureDuration,
		BoundaryDate:   parsedBoundary,
		OutputDir:      *outDir,
		DataFilePath:   dataFilePath,
		Reader:         os.Stdin,
		Writer:         os.Stdout,
	}

	n, err := core.NewNewsletter(c.OrganizationID, c.TeamIDs)
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, statErr := os.Stat(dataFilePath); statErr != nil {
		// No data file found — fetch live from USTA.
		if err := n.Generate(ctx); err != nil {
			fmt.Println(err)
			return
		}
	} else {
		slog.Info("data file found, will load instead of fetching from USTA", "path", dataFilePath)
	}

	data, err := formatters.Prepare(n, fmtCfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := c.RecentFormatter.FormatRecent(data, fmtCfg); err != nil {
		fmt.Println(err)
		return
	}

	if err := c.UpcomingFormatter.FormatUpcoming(data, fmtCfg); err != nil {
		fmt.Println(err)
		return
	}

	if err := data.Save(); err != nil {
		fmt.Println(err)
		return
	}

	slog.Info("done")
}
