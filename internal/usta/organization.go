package usta

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	organizationURL = "https://leagues.ustanorcal.com/organization.asp?id=%d"
)

var shortNameTranslations = map[string]string{
	"BCC":   "Courtside",
	"SMTC":  "Sunnyvale",
	"VG&CC": "Villages",
}

// Organization represents a USTA NorCal organization.
type Organization struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Teams []*Team `json:"teams"`

	doc *goquery.Document
}

// LoadOrganization loads the organization details for the given organization ID.
func LoadOrganization(id int) (*Organization, error) {
	cacheKey := fmt.Sprintf("org:%d", id)

	// Use singleflight to deduplicate concurrent requests
	result, err, _ := orgGroup.Do(cacheKey, func() (interface{}, error) {
		// Check cache first
		if cached, ok := orgCache.get(cacheKey); ok {
			return cached.(*Organization), nil
		}

		res, err := httpClient.Get(fmt.Sprintf(organizationURL, id))
		if err != nil {
			return nil, errors.Wrap(err, "could not fetch organization page")
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("error fetching organization page, code: %d", res.StatusCode)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "could not read organization page")
		}

		o := new(Organization)
		o.doc = doc
		o.ID = id
		o.Name = doc.Find("table tbody tr td font b").First().Text()

		// Store in cache
		orgCache.set(cacheKey, o)

		return o, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*Organization), nil
}

// LoadTeams loads teams for an organization.
func (o *Organization) LoadTeams() (*Organization, error) {
	var teamIDs []int

	o.doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		u, exists := sel.Attr("href")
		if !exists {
			return
		}

		if strings.HasPrefix(u, "teaminfo.asp?") {
			teamID, err := parseTeamID(u)
			if err != nil {
				return
			}

			teamIDs = append(teamIDs, teamID)
		}
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, teamID := range teamIDs {
		wg.Add(1)
		go func(teamID int) {
			defer wg.Done()
			t, err := LoadTeam(teamID)
			if err != nil {
				return
			}
			mu.Lock()
			o.Teams = append(o.Teams, t)
			mu.Unlock()
		}(teamID)
	}

	wg.Wait()
	return o, nil
}

func (o *Organization) ShortName() string {
	parts := strings.Split(strings.TrimSpace(o.Name), " ")

	var shortName string
	for _, part := range parts {
		if part == "AND" {
			continue
		}
		shortName += string(part[0])
	}

	if t, exists := shortNameTranslations[shortName]; exists {
		shortName = t
	}

	return shortName
}

func (o *Organization) Equals(ao *Organization) bool {
	return o.ID == ao.ID
}

func (o *Organization) Matches(past, future time.Duration) (pastMatches []Match, futureMatches []Match) {
	now := time.Now()

	start := now.Add(-1 * past)
	end := now.Add(future)

	for _, t := range o.Teams {
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
	orgTeamIDs := make(map[int]bool, len(o.Teams))
	for _, t := range o.Teams {
		orgTeamIDs[t.ID] = true
	}
	sort.Slice(futureMatches, func(i, j int) bool {
		if !futureMatches[i].Date.Equal(futureMatches[j].Date) {
			return futureMatches[i].Date.Before(futureMatches[j].Date)
		}
		// At the same time, show home matches first.
		return orgTeamIDs[futureMatches[i].HomeTeam.ID] && !orgTeamIDs[futureMatches[j].HomeTeam.ID]
	})

	return
}
