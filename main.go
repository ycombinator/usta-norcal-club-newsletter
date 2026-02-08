package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

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
  usta-norcal-club-newsletter                         Use default org (ASRC), console output
  usta-norcal-club-newsletter -org=300                Specify a different organization
  usta-norcal-club-newsletter -teams=123,456          Track additional teams by ID
  usta-norcal-club-newsletter -format=pdf             Generate PDF newsletter
  usta-norcal-club-newsletter help                    Show this help message
`)
}

func main() {
	c := internal.DefaultConfig()

	flag.Usage = usage
	orgID := flag.Int("org", c.OrganizationID, "USTA NorCal organization ID")
	teams := flag.String("teams", "", "comma-separated list of additional team IDs to track")
	format := flag.String("format", "console", "output format: console or pdf")

	// Handle "help" sub-command before flag.Parse
	if len(os.Args) > 1 && os.Args[1] == "help" {
		usage()
		return
	}

	flag.Parse()

	c.OrganizationID = *orgID

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

	switch *format {
	case "console":
		c.Formatter = formatters.NewConsoleFormatter()
	case "pdf":
		c.Formatter = formatters.NewPDFFormatter()
	default:
		fmt.Fprintf(os.Stderr, "unknown format: %s (use 'console' or 'pdf')\n", *format)
		os.Exit(1)
	}

	n, err := core.NewNewsletter(c.OrganizationID, c.TeamIDs)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := n.Generate(); err != nil {
		fmt.Println(err)
		return
	}

	fmtCfg := formatters.Config{
		OrganizationID: c.OrganizationID,
		PastDuration:   c.PastDuration,
		FutureDuration: c.FutureDuration,
	}
	if err := c.Formatter.Format(n, fmtCfg); err != nil {
		fmt.Println(err)
		return
	}
}
