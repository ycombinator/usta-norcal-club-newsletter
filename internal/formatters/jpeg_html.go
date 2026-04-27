package formatters

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"sort"
	"time"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

type RecentResultsData struct {
	OrgShortName string
	Rows         []ResultRow
	Footnotes    []string
}

type ResultRow struct {
	DayLabel     string
	IsWeekend    bool
	GenderEmoji  string
	Level        string
	DaytimeEmoji string
	OutcomeText  string
	IsWin        bool
	IsRainedOut  bool
	IsIncomplete bool
	LocatorEmoji string
	OpponentName string
	Tag          string
}

type UpcomingMatchesData struct {
	OrgShortName string
	Days         []CalendarDay
	MaxSlots     int
	Footnotes    []string
}

type CalendarDay struct {
	DayName   string
	Date      string
	IsWeekend bool
	Slots     []CalendarMatch
}

type CalendarMatch struct {
	Empty        bool
	LocatorEmoji string
	FootnoteMark string
	Time         string
	GenderEmoji  string
	Level        string
	DaytimeEmoji string
	OpponentName string
	Tag          string
}

func isWeekend(d time.Weekday) bool {
	return d == time.Saturday || d == time.Sunday
}

func resolveTeams(m usta.Match, org *usta.Organization) (ourTeam, opponent *usta.Team, isHome bool) {
	ctx := context.Background()
	m.HomeTeam.LoadOrganization(ctx)
	m.VisitingTeam.LoadOrganization(ctx)

	if m.HomeTeam.Organization.Equals(org) || m.HomeTeam.Extra {
		return m.HomeTeam, m.VisitingTeam, true
	}
	return m.VisitingTeam, m.HomeTeam, false
}

func opponentDisplayName(names *OrgNames, reader io.Reader, writer io.Writer, org *usta.Organization) string {
	return names.Resolve(reader, writer, org.Name)
}

func locationEmoji(isHome bool) string {
	if isHome {
		return "🏠"
	}
	return "🚗"
}

func matchTypeTag(mt MatchType) string {
	switch mt {
	case Playoff:
		return "playoff"
	case Sectionals:
		return "Sectionals"
	default:
		return ""
	}
}

func BuildRecentResultsData(org *usta.Organization, matches []AnnotatedMatch, names *OrgNames, reader io.Reader, writer io.Writer) RecentResultsData {
	data := RecentResultsData{
		OrgShortName: org.ShortName(),
	}

	footnoteSet := map[string]bool{}
	var currentDay string

	for _, am := range matches {
		m := am.Match
		ourTeam, opponent, isHome := resolveTeams(m, org)
		d := ourTeam.Display()

		opponent.LoadOrganization(context.Background())

		dayKey := m.Date.Format("2006-01-02")
		showLabel := dayKey != currentDay
		if showLabel {
			currentDay = dayKey
		}

		row := ResultRow{
			GenderEmoji:  d.GenderEmoji(),
			Level:        d.Level,
			DaytimeEmoji: d.DaytimeEmoji(),
			LocatorEmoji: locationEmoji(isHome),
			OpponentName: opponentDisplayName(names, reader, writer, opponent.Organization),
			Tag:          matchTypeTag(am.Annotation.MatchType),
			IsWeekend:    isWeekend(m.Date.Weekday()),
		}

		if showLabel {
			row.DayLabel = m.Date.Format("Mon 1/2")
		}

		if am.Annotation.RainedOut {
			row.IsRainedOut = true
		} else if am.Annotation.Score != "" {
			row.IsIncomplete = true
			row.OutcomeText = am.Annotation.Score
			if am.Annotation.Footnote != "" && !footnoteSet[am.Annotation.Footnote] {
				footnoteSet[am.Annotation.Footnote] = true
				data.Footnotes = append(data.Footnotes, am.Annotation.Footnote)
			}
		} else if am.Annotation.Footnote != "" {
			row.IsIncomplete = true
			if !footnoteSet[am.Annotation.Footnote] {
				footnoteSet[am.Annotation.Footnote] = true
				data.Footnotes = append(data.Footnotes, am.Annotation.Footnote)
			}
		} else if m.Outcome.WinningTeam != nil {
			m.Outcome.WinningTeam.LoadOrganization(context.Background())
			if m.Outcome.WinningTeam.Organization.Equals(ourTeam.Organization) || m.Outcome.WinningTeam == ourTeam {
				row.IsWin = true
				row.OutcomeText = fmt.Sprintf("won %d-%d", m.Outcome.WinnerPoints, m.Outcome.LoserPoints)
			} else {
				row.OutcomeText = fmt.Sprintf("lost %d-%d", m.Outcome.LoserPoints, m.Outcome.WinnerPoints)
			}
		}

		data.Rows = append(data.Rows, row)
	}

	return data
}

func BuildUpcomingMatchesData(org *usta.Organization, matches []usta.Match, names *OrgNames, locationOverrides map[int]string, reader io.Reader, writer io.Writer) UpcomingMatchesData {
	data := UpcomingMatchesData{
		OrgShortName: org.ShortName(),
	}

	if len(matches) == 0 {
		return data
	}

	firstDate := matches[0].Date
	monday := firstDate
	for monday.Weekday() != time.Monday {
		monday = monday.AddDate(0, 0, -1)
	}

	days := make([]CalendarDay, 7)

	for i := 0; i < 7; i++ {
		d := monday.AddDate(0, 0, i)
		days[i] = CalendarDay{
			DayName:   d.Format("Mon"),
			Date:      d.Format("1/2"),
			IsWeekend: isWeekend(d.Weekday()),
		}
	}

	type timedMatch struct {
		sortKey time.Time
		match   CalendarMatch
	}
	timedByDay := make([][]timedMatch, 7)

	superscripts := []string{"¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹"}
	footnoteIndex := map[string]int{}

	for i, m := range matches {
		dayIdx := int(m.Date.Weekday()) - int(time.Monday)
		if dayIdx < 0 {
			dayIdx += 7
		}
		if dayIdx >= 7 {
			continue
		}

		ourTeam, opponent, isHome := resolveTeams(m, org)
		d := ourTeam.Display()
		opponent.LoadOrganization(context.Background())

		cm := CalendarMatch{
			LocatorEmoji: locationEmoji(isHome),
			Time:         formatMatchTime(m.Date),
			GenderEmoji:  d.GenderEmoji(),
			Level:        d.Level,
			DaytimeEmoji: d.DaytimeEmoji(),
			OpponentName: opponentDisplayName(names, reader, writer, opponent.Organization),
		}

		if loc, ok := locationOverrides[i]; ok {
			idx, exists := footnoteIndex[loc]
			if !exists {
				idx = len(data.Footnotes)
				footnoteIndex[loc] = idx
				mark := superscripts[idx%len(superscripts)]
				data.Footnotes = append(data.Footnotes, mark+" at "+loc)
			}
			cm.FootnoteMark = superscripts[idx%len(superscripts)]
		}

		timedByDay[dayIdx] = append(timedByDay[dayIdx], timedMatch{sortKey: m.Date, match: cm})
	}

	type dayLayout struct {
		morning []CalendarMatch
		evening []CalendarMatch
	}
	layouts := make([]dayLayout, 7)

	for i := range timedByDay {
		sort.Slice(timedByDay[i], func(a, b int) bool {
			return timedByDay[i][a].sortKey.Before(timedByDay[i][b].sortKey)
		})
		for _, tm := range timedByDay[i] {
			if tm.sortKey.Hour() >= 16 {
				layouts[i].evening = append(layouts[i].evening, tm.match)
			} else {
				layouts[i].morning = append(layouts[i].morning, tm.match)
			}
		}
	}

	maxSlots := 0
	for _, l := range layouts {
		if n := len(l.morning) + len(l.evening); n > maxSlots {
			maxSlots = n
		}
	}

	for i := range days {
		days[i].Slots = make([]CalendarMatch, maxSlots)
		for j := range days[i].Slots {
			days[i].Slots[j] = CalendarMatch{Empty: true}
		}
		for j, cm := range layouts[i].morning {
			days[i].Slots[j] = cm
		}
		eveningStart := maxSlots - len(layouts[i].evening)
		for j, cm := range layouts[i].evening {
			days[i].Slots[eveningStart+j] = cm
		}
	}

	data.Days = days
	data.MaxSlots = maxSlots
	return data
}

func formatMatchTime(t time.Time) string {
	hour := t.Hour()
	minute := t.Minute()

	if hour == 0 && minute == 0 {
		return ""
	}

	period := "am"
	if hour >= 12 {
		period = "pm"
	}
	displayHour := hour % 12
	if displayHour == 0 {
		displayHour = 12
	}

	if minute == 0 {
		return fmt.Sprintf("%d%s", displayHour, period)
	}
	return fmt.Sprintf("%d:%02d%s", displayHour, minute, period)
}

const recentResultsHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body {
    font-family: 'Marker Felt', cursive;
    margin: 0;
    padding: 20px 24px;
    display: inline-block;
    white-space: nowrap;
  }
  .title {
    font-size: 28px;
    font-weight: bold;
    text-align: center;
    margin-bottom: 4px;
  }
  .subtitle {
    font-size: 22px;
    font-weight: bold;
    text-align: center;
    margin-bottom: 16px;
  }
  table { border-collapse: collapse; }
  td { padding: 4px 10px; vertical-align: middle; white-space: nowrap; font-size: 20px; }
  .day-label { font-weight: bold; font-style: italic; }
  .weekend { color: red; }
  .outcome { text-align: center; }
  .win { font-weight: bold; }
  .loss { font-style: italic; color: #999; }
  .rainedout { text-align: center; }
  .tag { background-color: yellow; padding: 1px 6px; border-radius: 4px; font-style: italic; }
  .footnotes { font-size: 16px; font-style: italic; text-align: right; margin-top: 10px; color: #666; }
  .team-col { font-size: 22px; font-weight: bold; }
  .opponent { font-weight: bold; }
</style>
</head>
<body>
  <div class="title">🏆🎾 {{.OrgShortName}} plays USTA league 🎾🏆</div>
  <div class="subtitle">Recent Results</div>
  <table>
    {{range .Rows}}
    <tr>
      <td class="day-label {{if .IsWeekend}}weekend{{end}}">{{.DayLabel}}</td>
      <td class="team-col">{{.GenderEmoji}}{{.Level}}{{.DaytimeEmoji}}</td>
      <td class="outcome {{if .IsRainedOut}}rainedout{{else if .IsWin}}win{{else}}loss{{end}}">{{if .IsRainedOut}}🌧️{{else if .IsIncomplete}}{{.OutcomeText}}*{{else}}{{.OutcomeText}}{{end}}</td>
      <td>{{.LocatorEmoji}}</td>
      <td class="opponent">{{.OpponentName}}</td>
      {{if .Tag}}<td><span class="tag">{{.Tag}}</span></td>{{end}}
    </tr>
    {{end}}
  </table>
  {{if .Footnotes}}<div class="footnotes">{{range .Footnotes}}<div>* {{.}}</div>{{end}}</div>{{end}}
</body>
</html>`

const upcomingMatchesHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body {
    font-family: 'Marker Felt', cursive;
    margin: 0;
    padding: 20px 24px;
    display: inline-block;
    white-space: nowrap;
  }
  .title {
    font-size: 28px;
    font-weight: bold;
    text-align: center;
    margin-bottom: 4px;
  }
  .subtitle {
    font-size: 22px;
    font-weight: bold;
    text-align: center;
    margin-bottom: 16px;
  }
  table {
    border-collapse: collapse;
  }
  th {
    font-size: 18px;
    font-weight: bold;
    font-style: italic;
    padding: 6px 12px;
    border: 1px solid #ccc;
    text-align: center;
    white-space: nowrap;
  }
  td {
    border: 1px solid #ccc;
    padding: 6px 10px;
    vertical-align: top;
    font-size: 17px;
    min-width: 100px;
  }
  .weekend { color: red; }
  .match-entry { margin-bottom: 2px; }
  .match-time { font-weight: bold; }
  .match-opponent { font-weight: bold; }
  .tag { background-color: yellow; padding: 1px 4px; border-radius: 4px; font-style: italic; font-size: 14px; }
  .footnotes { font-size: 14px; font-style: italic; text-align: right; margin-top: 10px; color: #666; }
  .empty-cell { }
</style>
</head>
<body>
  <div class="title">🏆🎾 {{.OrgShortName}} plays USTA league 🎾🏆</div>
  <div class="subtitle">Upcoming Matches</div>
  <table>
    <tr>
      {{range .Days}}<th class="{{if .IsWeekend}}weekend{{end}}">{{.DayName}}<br>{{.Date}}</th>{{end}}
    </tr>
    {{range $slot := Slots .MaxSlots}}
    <tr>
      {{range $.Days}}
      <td>
        {{with index .Slots $slot}}
        {{if not .Empty}}
        <div class="match-entry">
          {{if .Tag}}<span class="tag">{{.Tag}}</span><br>{{end}}
          {{.LocatorEmoji}}{{.FootnoteMark}} <span class="match-time">{{.Time}}</span><br>
          {{.GenderEmoji}} {{.Level}}{{.DaytimeEmoji}}<br>
          <span class="match-opponent">{{.OpponentName}}</span>
        </div>
        {{end}}
        {{end}}
      </td>
      {{end}}
    </tr>
    {{end}}
  </table>
  {{if .Footnotes}}<div class="footnotes">{{range .Footnotes}}<div>{{.}}</div>{{end}}</div>{{end}}
</body>
</html>`

var templateFuncs = template.FuncMap{
	"Slots": func(n int) []int {
		s := make([]int, n)
		for i := range s {
			s[i] = i
		}
		return s
	},
}

func RenderRecentResultsHTML(data RecentResultsData) (string, error) {
	tmpl, err := template.New("recent").Parse(recentResultsHTML)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderUpcomingMatchesHTML(data UpcomingMatchesData) (string, error) {
	tmpl, err := template.New("upcoming").Funcs(templateFuncs).Parse(upcomingMatchesHTML)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
