package core

import (
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
	for _, t := range n.org.Teams {
		wg.Add(1)
		go func(t *usta.Team) {
			t.LoadMatches()
			wg.Done()
		}(t)
	}

	wg.Wait()
	return nil
}
