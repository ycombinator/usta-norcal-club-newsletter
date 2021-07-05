package internal

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type Newsletter struct {
	cfg Config

	org Organization
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
	n.org = *org

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

	for _, t := range n.org.Teams {
		for _, m := range t.Matches {
			if m.Date.After(now) && m.Date.Before(now.Add(n.cfg.FutureDuration)) {
				futureMatches = append(futureMatches, m)
			} else if m.Date.Before(now) && m.Date.After(now.Add(-1*n.cfg.PastDuration)) {
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

	str.WriteString("Recent matches (home team shown first):\n")
	for _, m := range pastMatches {
		str.WriteString(m.String())
		str.WriteString("\n")
	}
	str.WriteString("\n")

	str.WriteString("Upcoming matches (home team shown first):\n")
	for _, m := range futureMatches {
		str.WriteString(m.String())
		str.WriteString("\n")
	}
	str.WriteString("\n")

	return str.String()
}
