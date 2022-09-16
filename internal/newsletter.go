package internal

import (
	"sync"
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

func (n Newsletter) Config() Config {
	return n.cfg
}

func (n Newsletter) Organization() *Organization {
	return n.org
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
