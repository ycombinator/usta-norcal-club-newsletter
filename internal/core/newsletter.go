package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

type Newsletter struct {
	orgID   int
	teamIDs []int
	org     *usta.Organization
}

func NewNewsletter(orgID int, teamIDs []int) (*Newsletter, error) {
	n := new(Newsletter)
	n.orgID = orgID
	n.teamIDs = teamIDs

	return n, nil
}

func (n Newsletter) Organization() *usta.Organization {
	return n.org
}

func (n *Newsletter) Generate(ctx context.Context) error {
	slog.Info("loading organization", "org_id", n.orgID)
	org, err := usta.LoadOrganization(ctx, n.orgID)
	if err != nil {
		return err
	}
	n.org = org
	slog.Info("loaded organization", "org_id", n.orgID, "name", org.Name)

	slog.Info("loading teams for organization")
	if _, err = n.org.LoadTeams(ctx); err != nil {
		return err
	}
	slog.Info("loaded teams", "count", len(n.org.Teams))

	if len(n.teamIDs) > 0 {
		slog.Info("loading extra teams", "team_ids", n.teamIDs)
		var wg sync.WaitGroup
		var mu sync.Mutex
		var extraTeams []*usta.Team
		var errs []error
		for _, id := range n.teamIDs {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				t, err := usta.LoadTeam(ctx, id)
				if err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
					return
				}
				t.Extra = true
				mu.Lock()
				extraTeams = append(extraTeams, t)
				mu.Unlock()
			}(id)
		}
		wg.Wait()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if len(errs) > 0 {
			return fmt.Errorf("failed to load %d extra team(s)", len(errs))
		}
		n.org.Teams = append(n.org.Teams, extraTeams...)
		slog.Info("loaded extra teams", "count", len(extraTeams))
	}

	slog.Info("loading matches for all teams", "team_count", len(n.org.Teams))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	for _, t := range n.org.Teams {
		wg.Add(1)
		go func(t *usta.Team) {
			defer wg.Done()
			if _, err := t.LoadMatches(ctx); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(t)
	}

	wg.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to load matches for %d team(s)", len(errs))
	}

	totalMatches := 0
	for _, t := range n.org.Teams {
		totalMatches += len(t.Matches)
	}
	slog.Info("loaded all matches", "total_matches", totalMatches)

	return nil
}
