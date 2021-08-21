package internal

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
)

type Newsletter struct {
	cfg Config

	org *Organization
}

func NewNewsletter(cfg Config) (*Newsletter, error) {
	n := new(Newsletter)
	n.cfg = cfg

	return n, nil
}

func (n *Newsletter) Generate() error {
	org, err := LoadOrganization(n.cfg.OrganizationID)
	if err != nil {
		return err
	}
	n.org = org

	if _, err = n.org.LoadTeams(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, t := range n.org.Teams {
		wg.Add(1)
		go func(t *Team) {
			t.LoadMatches()
			wg.Done()
		}(t)
	}

	wg.Wait()
	return nil
}

func (n *Newsletter) String() string {
	var pastMatches, futureMatches []Match
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	start := now.Add(-1*n.cfg.PastDuration)
	end := now.Add(n.cfg.FutureDuration)

	for _, t := range n.org.Teams {
		for _, m := range t.Matches {
			if (m.Date.Equal(now) || m.Date.After(now)) && m.Date.Before(end) {
				futureMatches = append(futureMatches, m)
			} else if m.Date.Before(now) && (m.Date.Equal(start) || m.Date.After(start)) {
				pastMatches = append(pastMatches, m)
			}
		}
	}

	// Sort matches
	sort.Slice(pastMatches, func(i, j int) bool {
		return pastMatches[i].Date.Before(pastMatches[j].Date)
	})
	sort.Slice(futureMatches, func(i, j int) bool {
		return futureMatches[i].Date.Before(futureMatches[j].Date)
	})

	var str strings.Builder

	if len(pastMatches) > 0 {
		str.WriteString("Recent matches:\n")
		table := tablewriter.NewWriter(&str)
		table.SetAutoWrapText(false)
		for _, m := range pastMatches {
			date, first, outcome, locator, second := m.ForOrganization(n.org)
			table.Append([]string{
				date.Format("Mon, Jan 02"),
				first,
				outcome,
				locator + " " + second,
			})
		}
		table.Render()
		str.WriteString("\n")
	}

	if len(futureMatches) > 0 {
		str.WriteString("Upcoming matches:\n")
		table := tablewriter.NewWriter(&str)
		table.SetAutoWrapText(false)
		for _, m := range futureMatches {
			date, first, _, locator, second := m.ForOrganization(n.org)
			table.Append([]string{
				date.Format("Mon, Jan 02"),
				first,
				locator + " " + second,
			})
		}
		table.Render()
		str.WriteString("\n")
	}

	return str.String()
}
