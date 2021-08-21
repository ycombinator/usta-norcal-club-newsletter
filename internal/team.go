package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	teamURL = "https://www.ustanorcal.com/teaminfo.asp?id=%d"
)

var (
	tz, _ = time.LoadLocation("America/Los_Angeles")
)

// Team represents a USTA NorCal team.
type Team struct {
	ID           int           `json:"id"`
	Organization *Organization `json:"organization"`
	Name         string        `json:"name"`
	FriendlyName string        `json:"friendly_name`
	Matches      []Match       `json:"matches`
	Players      []Player      `json:"players"`

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
	if t.Organization != nil {
		return t, nil
	}

	var orgID int

	t.doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		if orgID != 0 {
			return
		}

		u, exists := sel.Attr("href")
		if !exists {
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

	return t, nil
}

// LoadMatches loads the matches for a team.
func (t *Team) LoadMatches() (*Team, error) {
	if t.Matches != nil {
		return t, nil
	}

	t.doc.Find("table tbody tr td table tbody tr").Each(func(i int, sel *goquery.Selection) {
		bgcolor, exists := sel.Attr("bgcolor")
		if !exists {
			return
		}

		if bgcolor != "white" && bgcolor != "#D2D2FF" {
			return
		}

		// Parse match date
		cells := sel.Find("td")
		if cells.Length() < 2 {
			return
		}

		c := cells.Get(2).FirstChild
		if c.NextSibling != nil {
			c = c.NextSibling.FirstChild
		}

		v := strings.TrimSpace(c.Data)
		dt, err := time.ParseInLocation("01/02/06", v, tz)
		if err != nil {
			return
		}

		// Parse opposing team ID
		v = sel.Find("td").Get(5).FirstChild.Attr[0].Val
		teamID, err := parseTeamID(v)
		if err != nil {
			return
		}

		o, err := LoadTeam(teamID)
		if err != nil {
			return
		}

		// Parse location (home or away)
		location := sel.Find("td").Get(6).FirstChild.Data

		var homeTeam, visitingTeam *Team
		if location == "Home" {
			homeTeam = t
			visitingTeam = o
		} else {
			homeTeam = o
			visitingTeam = t
		}

		m := Match{
			Date:         dt,
			HomeTeam:     homeTeam,
			VisitingTeam: visitingTeam,
		}

		// Parse outcome
		v = sel.Find("td").Get(7).FirstChild.Data
		verb, winnerPoints, loserPoints, err := parseOutcome(v)
		if err != nil {
			return
		}

		if verb != "" {
			var winningTeam *Team
			if verb == "Won" {
				winningTeam = t
			} else {
				winningTeam = o
			}

			outcome := struct {
				WinningTeam  *Team
				WinnerPoints int
				LoserPoints  int
			}{
				WinningTeam:  winningTeam,
				WinnerPoints: winnerPoints,
				LoserPoints:  loserPoints,
			}

			m.Outcome = outcome
		}

		t.Matches = append(t.Matches, m)
	})

	return t, nil
}

func (t *Team) ShortName() string {
	var shortName = t.Name

	// Strip current year out of short name
	year := time.Now().Format("2006")
	shortName = strings.Replace(shortName, year+" ", "", -1)

	// Abbreviate "& Over"
	shortName = strings.Replace(shortName, " & Over", "+", -1)

	return shortName
}

func parseTeamID(u string) (int, error) {
	pu, err := url.Parse(u)
	if err != nil {
		return 0, fmt.Errorf("could not parse team URL: %w", err)
	}

	v := pu.Query().Get("id")
	teamID, err := strconv.ParseInt(v, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("could not parse team ID from team URL: %w", err)
	}

	return int(teamID), nil
}

func parseOutcome(outcome string) (string, int, int, error) {
	outcome = strings.TrimSpace(outcome)
	parts := strings.Split(outcome, " ")
	if len(parts) != 2 {
		return "", 0, 0, nil
	}

	verb := parts[0]
	points := strings.Split(parts[1], "-")

	points1, err := strconv.ParseInt(points[0], 10, 0)
	if err != nil {
		return "", 0, 0, err
	}

	points2, err := strconv.ParseInt(points[1], 10, 0)
	if err != nil {
		return "", 0, 0, err
	}

	var winnerPoints, loserPoints int64
	if points1 > points2 {
		winnerPoints = points1
		loserPoints = points2
	} else {
		winnerPoints = points2
		loserPoints = points1
	}

	return verb, int(winnerPoints), int(loserPoints), nil
}
