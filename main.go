package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/formatters"
)

func main() {
	c := internal.DefaultConfig()

	if len(os.Args) > 1 {
		oID, err := strconv.ParseInt(os.Args[1], 10, 0)
		if err != nil {
			fmt.Println("organization ID must be an integer")
			return
		}

		c.OrganizationID = int(oID)
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
