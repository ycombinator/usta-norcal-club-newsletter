package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/formatters"
)

func main() {
	c := internal.DefaultConfig()

	orgID := flag.Int("org", c.OrganizationID, "USTA NorCal organization ID")
	format := flag.String("format", "console", "output format: console or pdf")
	flag.Parse()

	c.OrganizationID = *orgID

	switch *format {
	case "console":
		c.Formatter = formatters.NewConsoleFormatter()
	case "pdf":
		c.Formatter = formatters.NewPDFFormatter()
	default:
		fmt.Fprintf(os.Stderr, "unknown format: %s (use 'console' or 'pdf')\n", *format)
		os.Exit(1)
	}

	n, err := core.NewNewsletter(c.OrganizationID)
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
