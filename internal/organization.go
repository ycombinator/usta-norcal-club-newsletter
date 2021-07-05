package internal

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	organizationURL = "https://www.ustanorcal.com/organization.asp?id=%d"
)

var shortNameTranslations = map[string]string{
	"BCC":  "Courtside",
	"SMTC": "Sunnyvale",
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
	res, err := http.Get(fmt.Sprintf(organizationURL, id))
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch organization page")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.Wrapf(err, "error fetching organization page, code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read organization page")
	}

	o := new(Organization)
	o.doc = doc
	o.ID = id
	o.Name = doc.Find("table tbody tr td font b").First().Text()

	return o, nil
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
	for _, teamID := range teamIDs {
		wg.Add(1)
		go func(teamID int) {
			t, _ := LoadTeam(teamID)
			o.Teams = append(o.Teams, t)
			wg.Done()
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
