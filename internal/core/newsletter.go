package core

import (
	"fmt"
	"sync"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

type Newsletter struct {
	orgID int
	org   *usta.Organization
}

func NewNewsletter(orgID int) (*Newsletter, error) {
	n := new(Newsletter)
	n.orgID = orgID

	return n, nil
}

func (n Newsletter) Organization() *usta.Organization {
	return n.org
}

func (n *Newsletter) Generate() error {
	org, err := usta.LoadOrganization(n.orgID)
	if err != nil {
		return err
	}
	n.org = org

	if _, err = n.org.LoadTeams(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	for _, t := range n.org.Teams {
		wg.Add(1)
		go func(t *usta.Team) {
			defer wg.Done()
			if _, err := t.LoadMatches(); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(t)
	}

	wg.Wait()
	if len(errs) > 0 {
		return fmt.Errorf("failed to load matches for %d team(s)", len(errs))
	}
	return nil
}
