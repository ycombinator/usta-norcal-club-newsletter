package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	teamURL = "https://www.ustanorcal.com/teaminfo.asp?id=%d"
)

// Team represents a USTA NorCal team.
type Team struct {
	ID int `json:"id"`
	Organization *Organization `json:"organization"`
	Name string `json:"name"`
	FriendlyName string `json:"friendly_name`
	Matches []Match `json:"matches`
	Players []Player `json:"players"`

	doc *goquery.Document
}

// LoadTeam loads a team's information for the given team ID.
func LoadTeam(id int) (*Team, error) {
  res, err := http.Get(fmt.Sprintf(teamURL, id))
  if err != nil {
		return nil, errors.Wrap(err, "could not fetch team page")
  }
  defer res.Body.Close()
  if res.StatusCode != 200 {
		return nil, errors.Wrapf(err, "error fetching team page, code: %d", res.StatusCode)
  }

	doc, err := goquery.NewDocumentFromReader(res.Body)
  if err != nil {
		return nil, errors.Wrap(err, "could not read team page")
  }

	t := new(Team)
	t.doc = doc
	t.ID = id

	t.Name = doc.Find("table tbody tr td b").First().Text()

	return t, nil
}

// LoadOrganization loads the organization for a team.
func (t *Team) LoadOrganization() (*Team, error) {
	var orgID int

	t.doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		if orgID != 0 {
			return
		}

		u, exists := sel.Attr("href")
		if !exists{ 
			return
		}

		if strings.HasPrefix(u, "organization.asp?") {
			pu, err := url.Parse(u)
			if err != nil {
				return
			}

			v := pu.Query().Get("id")
			oID, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				return
			}

			orgID = int(oID)
		}
	})

	o, err := LoadOrganization(orgID)
	if err != nil {
		return nil, errors.Wrapf(err, "could not load organization for team ID = %d", t.ID)
	}
	t.Organization = o

	return t,nil
}