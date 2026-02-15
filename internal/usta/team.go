package usta

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	teamURL = "https://leagues.ustanorcal.com/teaminfo.asp?id=%d"
)

var (
	tz, _     = time.LoadLocation("America/Los_Angeles")
	timeRegex = regexp.MustCompile(`at[^\d]+(\d+):(\d\d)\s+([aApP]M)`)
)

// Team represents a USTA NorCal team.
type Team struct {
	ID           int           `json:"id"`
	Organization *Organization `json:"organization"`
	Name         string        `json:"name"`
	Matches      []Match       `json:"matches"`

	doc *goquery.Document
}

// LoadTeam loads a team's information for the given team ID.
func LoadTeam(id int) (*Team, error) {
	cacheKey := fmt.Sprintf("team:%d", id)

	// Use singleflight to deduplicate concurrent requests
	result, err, _ := teamGroup.Do(cacheKey, func() (interface{}, error) {
		// Check cache first
		if cached, ok := teamCache.get(cacheKey); ok {
			return cached.(*Team), nil
		}

		res, err := httpClient.Get(fmt.Sprintf(teamURL, id))
		if err != nil {
			return nil, errors.Wrap(err, "could not fetch team page")
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("error fetching team page, code: %d", res.StatusCode)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "could not read team page")
		}

		t := new(Team)
		t.doc = doc
		t.ID = id

		t.Name = doc.Find("table tbody tr td b").First().Text()

		// Store in cache
		teamCache.set(cacheKey, t)

		return t, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*Team), nil
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

	// First pass: collect all match data and opposing team IDs
	type matchData struct {
		date         time.Time
		teamID       int
		location     string
		outcomeVerb  string
		winnerPoints int
		loserPoints  int
	}

	var matchDataList []matchData
	var opposingTeamIDs []int

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

		// Parse match times
		c = cells.Get(4).FirstChild

		v = strings.TrimSpace(c.Data)
		hour, minute, err := parseTime(v)
		if err != nil {
			return
		}
		if hour > 0 {
			dt = time.Date(dt.Year(), dt.Month(), dt.Day(), hour, minute, 0, 0, dt.Location())
		}

		// Parse opposing team ID
		v = cells.Get(5).FirstChild.Attr[0].Val
		teamID, err := parseTeamID(v)
		if err != nil {
			return
		}

		// Parse location (home or away)
		location := sel.Find("td").Get(6).FirstChild.Data

		// Parse outcome
		v = sel.Find("td").Get(7).FirstChild.Data
		verb, winnerPoints, loserPoints, err := parseOutcome(v)
		if err != nil {
			return
		}

		matchDataList = append(matchDataList, matchData{
			date:         dt,
			teamID:       teamID,
			location:     location,
			outcomeVerb:  verb,
			winnerPoints: winnerPoints,
			loserPoints:  loserPoints,
		})
		opposingTeamIDs = append(opposingTeamIDs, teamID)
	})

	// Second pass: load all opposing teams in parallel
	type teamResult struct {
		team *Team
		err  error
		idx  int
	}

	teamChan := make(chan teamResult, len(opposingTeamIDs))

	for idx, teamID := range opposingTeamIDs {
		go func(idx, teamID int) {
			team, err := LoadTeam(teamID)
			teamChan <- teamResult{team: team, err: err, idx: idx}
		}(idx, teamID)
	}

	// Collect results
	opposingTeams := make([]*Team, len(opposingTeamIDs))
	for range opposingTeamIDs {
		result := <-teamChan
		if result.err != nil {
			continue // Skip teams that failed to load
		}
		opposingTeams[result.idx] = result.team
	}

	// Third pass: build matches with loaded teams
	for idx, md := range matchDataList {
		o := opposingTeams[idx]
		if o == nil {
			continue // Skip if team failed to load
		}

		var homeTeam, visitingTeam *Team
		if md.location == "Home" {
			homeTeam = t
			visitingTeam = o
		} else {
			homeTeam = o
			visitingTeam = t
		}

		m := Match{
			Date:         md.date,
			HomeTeam:     homeTeam,
			VisitingTeam: visitingTeam,
		}

		if md.outcomeVerb != "" {
			var winningTeam *Team
			if md.outcomeVerb == "Won" {
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
				WinnerPoints: md.winnerPoints,
				LoserPoints:  md.loserPoints,
			}

			m.Outcome = outcome
		}

		t.Matches = append(t.Matches, m)
	}

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
func parseTime(u string) (int, int, error) {
	u = strings.TrimSpace(u)
	if u == "" {
		return 0, 0, nil
	}

	parts := timeRegex.FindStringSubmatch(u)
	if len(parts) < 4 {
		return 0, 0, nil
	}
	hour, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		return 0, 0, err
	}

	minute, err := strconv.Atoi(string(parts[2]))
	if err != nil {
		return 0, 0, err
	}

	ampm := strings.ToLower(string(parts[3]))
	if ampm == "pm" && hour < 12 {
		hour += 12
	} else if ampm == "am" && hour == 12 {
		hour = 0
	}

	return hour, minute, nil
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
