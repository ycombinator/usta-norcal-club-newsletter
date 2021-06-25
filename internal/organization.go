package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	organizationURL = "https://www.ustanorcal.com/organization.asp?id=%d"
)

// Organization represents a USTA NorCal organization.
type Organization struct {
	ID int `json:"id"`
	Name string `json:"name"`
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
		if !exists{ 
			return
		}

		if strings.HasPrefix(u, "teaminfo.asp?") {
			pu, err := url.Parse(u)
			if err != nil {
				return
			}

			v := pu.Query().Get("id")
			teamID, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				return
			}

			teamIDs = append(teamIDs, int(teamID))
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
}

// String returns the string representation of an organization.
func (o *Organization) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Name: %s\n", o.Name))

	if len(o.Teams) > 0 {
		sb.WriteString("Teams:\n")
		for _, t := range o.Teams {
			sb.WriteString(fmt.Sprintf("- %s\n", t.Name))
		}
	}

	return sb.String()
}