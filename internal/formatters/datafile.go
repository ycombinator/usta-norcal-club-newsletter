package formatters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ycombinator/usta-norcal-club-newsletter/internal/usta"
)

// DataFile is the intermediate JSON data saved alongside output files.
// Edit any "date" field to correct a wrong date, then re-run to regenerate the report.
type DataFile struct {
	OrgShortName  string              `json:"org_short_name"`
	PastMatches   []PastMatchRecord   `json:"past_matches"`
	FutureMatches []FutureMatchRecord `json:"future_matches"`
}

// PastMatchRecord is a human-editable record for a single past match.
type PastMatchRecord struct {
	Date         string `json:"date"`                    // YYYY-MM-DD; change to correct wrong dates
	GenderEmoji  string `json:"gender_emoji"`
	Level        string `json:"level"`
	Superscript  string `json:"superscript,omitempty"`   // team suffix: "A", "B", etc.
	IsHome       bool   `json:"is_home"`
	Opponent     string `json:"opponent"`
	IsWin        bool   `json:"is_win,omitempty"`
	IsRainedOut  bool   `json:"is_rained_out,omitempty"`
	IsIncomplete bool   `json:"is_incomplete,omitempty"`
	OutcomeText  string `json:"outcome_text,omitempty"`  // "won 2-1" or partial score
	Footnote     string `json:"footnote,omitempty"`
	MatchType    string `json:"match_type,omitempty"` // "regular", "playoff", "sectionals"
}

// FutureMatchRecord is a human-editable record for a single upcoming match.
type FutureMatchRecord struct {
	Date         string `json:"date"`                    // YYYY-MM-DD; change to correct wrong dates
	Time         string `json:"time,omitempty"`          // HH:MM in 24-hour format
	GenderEmoji  string `json:"gender_emoji"`
	Level        string `json:"level"`
	Superscript  string `json:"superscript,omitempty"`
	IsHome       bool   `json:"is_home"`
	Opponent     string `json:"opponent"`
	LocationNote string `json:"location_note,omitempty"` // alternate location for away extra-team matches
	MatchType    string `json:"match_type,omitempty"`    // "regular", "playoff", "sectionals"
}

// NewDataFile builds a DataFile from a PreparedData populated via live USTA data.
func NewDataFile(data *PreparedData, reader io.Reader, writer io.Writer) *DataFile {
	df := &DataFile{
		OrgShortName: data.Org.ShortName(),
	}
	for _, am := range data.PastMatches {
		df.PastMatches = append(df.PastMatches, buildPastRecord(am, data.Org, data.OrgNames, reader, writer))
	}
	for i, m := range data.FutureMatches {
		loc := data.LocationOverrides[i]
		df.FutureMatches = append(df.FutureMatches, buildFutureRecord(m, data.Org, data.OrgNames, loc, reader, writer))
	}
	return df
}

func buildPastRecord(am AnnotatedMatch, org *usta.Organization, names *OrgNames, reader io.Reader, writer io.Writer) PastMatchRecord {
	m := am.Match
	ourTeam, opponent, isHome := resolveTeams(m, org)
	d := ourTeam.Display()
	opponent.LoadOrganization(context.Background())

	rec := PastMatchRecord{
		Date:        m.Date.Format("2006-01-02"),
		GenderEmoji: d.GenderEmoji(),
		Level:       d.Level,
		Superscript: suffixForTeam(org, ourTeam),
		IsHome:      isHome,
		Opponent:    opponentDisplayName(names, reader, writer, opponent.Organization),
		MatchType:   matchTypeToString(am.Annotation.MatchType),
	}

	if am.Annotation.RainedOut {
		rec.IsRainedOut = true
	} else if am.Annotation.Score != "" {
		rec.IsIncomplete = true
		rec.OutcomeText = am.Annotation.Score
		rec.Footnote = am.Annotation.Footnote
	} else if am.Annotation.Footnote != "" {
		rec.IsIncomplete = true
		rec.Footnote = am.Annotation.Footnote
	} else if m.Outcome.WinningTeam != nil {
		m.Outcome.WinningTeam.LoadOrganization(context.Background())
		if m.Outcome.WinningTeam.Organization.Equals(ourTeam.Organization) || m.Outcome.WinningTeam == ourTeam {
			rec.IsWin = true
			rec.OutcomeText = fmt.Sprintf("won %d-%d", m.Outcome.WinnerPoints, m.Outcome.LoserPoints)
		} else {
			rec.OutcomeText = fmt.Sprintf("lost %d-%d", m.Outcome.LoserPoints, m.Outcome.WinnerPoints)
		}
	}

	return rec
}

func buildFutureRecord(m usta.Match, org *usta.Organization, names *OrgNames, locationNote string, reader io.Reader, writer io.Writer) FutureMatchRecord {
	ourTeam, opponent, isHome := resolveTeams(m, org)
	d := ourTeam.Display()
	opponent.LoadOrganization(context.Background())

	rec := FutureMatchRecord{
		Date:         m.Date.Format("2006-01-02"),
		GenderEmoji:  d.GenderEmoji(),
		Level:        d.Level,
		Superscript:  suffixForTeam(org, ourTeam),
		IsHome:       isHome,
		Opponent:     opponentDisplayName(names, reader, writer, opponent.Organization),
		LocationNote: locationNote,
	}
	if m.HasTime {
		rec.Time = m.Date.Format("15:04")
	}
	rec.MatchType = matchTypeToString(matchTypeFromString(m.MatchTypeHint))
	return rec
}

func matchTypeToString(mt MatchType) string {
	switch mt {
	case Playoff:
		return "playoff"
	case Sectionals:
		return "sectionals"
	default:
		return "regular"
	}
}

func matchTypeFromString(s string) MatchType {
	switch s {
	case "playoff":
		return Playoff
	case "sectionals":
		return Sectionals
	default:
		return RegularSeason
	}
}

// Save writes the DataFile to path as indented JSON, creating the parent
// directory if it doesn't already exist.
func (df *DataFile) Save(path string) error {
	b, err := json.MarshalIndent(df, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling data file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating data file directory: %w", err)
	}
	return os.WriteFile(path, b, 0644)
}

// LoadDataFile reads a DataFile from path.
func LoadDataFile(path string) (*DataFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var df DataFile
	if err := json.Unmarshal(b, &df); err != nil {
		return nil, fmt.Errorf("parsing data file %s: %w", path, err)
	}
	return &df, nil
}

// ToRecentResultsData builds display data from the data file records, re-sorted by date.
// Changing a "date" value in the JSON and re-running will move that match to the correct day.
func (df *DataFile) ToRecentResultsData() RecentResultsData {
	data := RecentResultsData{OrgShortName: df.OrgShortName}

	sorted := make([]PastMatchRecord, len(df.PastMatches))
	copy(sorted, df.PastMatches)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date < sorted[j].Date
	})

	footnoteSet := map[string]bool{}
	var prevDate string

	for _, rec := range sorted {
		t, _ := time.ParseInLocation("2006-01-02", rec.Date, time.Local)

		row := ResultRow{
			GenderEmoji:     rec.GenderEmoji,
			Level:           rec.Level,
			TeamSuperscript: teamSuperscript(rec.Superscript),
			LocatorEmoji:    locationEmoji(rec.IsHome),
			OpponentName:    rec.Opponent,
			Tag:             matchTypeTag(matchTypeFromString(rec.MatchType)),
			IsWeekend:       isWeekend(t.Weekday()),
			IsWin:           rec.IsWin,
			IsRainedOut:     rec.IsRainedOut,
			IsIncomplete:    rec.IsIncomplete,
			OutcomeText:     rec.OutcomeText,
		}

		if rec.Date != prevDate {
			row.DayLabel = t.Format("Mon 1/2")
			prevDate = rec.Date
		}

		if rec.Footnote != "" && !footnoteSet[rec.Footnote] {
			footnoteSet[rec.Footnote] = true
			data.Footnotes = append(data.Footnotes, rec.Footnote)
		}

		data.Rows = append(data.Rows, row)
	}

	return data
}

// ToUpcomingMatchesData builds a weekly calendar display from the data file records.
func (df *DataFile) ToUpcomingMatchesData() UpcomingMatchesData {
	data := UpcomingMatchesData{OrgShortName: df.OrgShortName}

	if len(df.FutureMatches) == 0 {
		return data
	}

	sorted := make([]FutureMatchRecord, len(df.FutureMatches))
	copy(sorted, df.FutureMatches)
	sort.Slice(sorted, func(i, j int) bool {
		di, ti := sorted[i].Date, sorted[i].Time
		dj, tj := sorted[j].Date, sorted[j].Time
		if di != dj {
			return di < dj
		}
		return ti < tj
	})

	firstDate, _ := time.ParseInLocation("2006-01-02", sorted[0].Date, time.Local)
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

	superscripts := []string{"¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹"}
	footnoteIndex := map[string]int{}

	type timedMatch struct {
		sortKey string
		isMorning bool
		match   CalendarMatch
	}
	timedByDay := make([][]timedMatch, 7)

	for _, rec := range sorted {
		d, _ := time.ParseInLocation("2006-01-02", rec.Date, time.Local)
		dayIdx := int(d.Weekday()) - int(time.Monday)
		if dayIdx < 0 {
			dayIdx += 7
		}
		if dayIdx >= 7 {
			continue
		}

		hour := 0
		if rec.Time != "" {
			fmt.Sscanf(rec.Time, "%d:", &hour)
		}

		cm := CalendarMatch{
			LocatorEmoji:    locationEmoji(rec.IsHome),
			Time:            dataFileMatchTime(rec.Time),
			GenderEmoji:     rec.GenderEmoji,
			Level:           rec.Level,
			TeamSuperscript: teamSuperscript(rec.Superscript),
			OpponentName:    rec.Opponent,
			Tag:             matchTypeTag(matchTypeFromString(rec.MatchType)),
		}

		if rec.LocationNote != "" {
			idx, exists := footnoteIndex[rec.LocationNote]
			if !exists {
				idx = len(data.Footnotes)
				footnoteIndex[rec.LocationNote] = idx
				mark := superscripts[idx%len(superscripts)]
				data.Footnotes = append(data.Footnotes, mark+" at "+rec.LocationNote)
			}
			cm.FootnoteMark = superscripts[idx%len(superscripts)]
		}

		sortKey := rec.Time
		if sortKey == "" {
			sortKey = "00:00"
		}
		timedByDay[dayIdx] = append(timedByDay[dayIdx], timedMatch{
			sortKey:   sortKey,
			isMorning: hour < 16,
			match:     cm,
		})
	}

	type dayLayout struct {
		morning []CalendarMatch
		evening []CalendarMatch
	}
	layouts := make([]dayLayout, 7)
	for i := range timedByDay {
		sort.Slice(timedByDay[i], func(a, b int) bool {
			return timedByDay[i][a].sortKey < timedByDay[i][b].sortKey
		})
		for _, tm := range timedByDay[i] {
			if tm.isMorning {
				layouts[i].morning = append(layouts[i].morning, tm.match)
			} else {
				layouts[i].evening = append(layouts[i].evening, tm.match)
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

// dataFileMatchTime converts "HH:MM" (24h) to "6pm" display format.
func dataFileMatchTime(t string) string {
	if t == "" {
		return ""
	}
	var h, m int
	fmt.Sscanf(t, "%d:%d", &h, &m)
	period := "am"
	if h >= 12 {
		period = "pm"
	}
	displayH := h % 12
	if displayH == 0 {
		displayH = 12
	}
	if m == 0 {
		return fmt.Sprintf("%d%s", displayH, period)
	}
	return fmt.Sprintf("%d:%02d%s", displayH, m, period)
}

// dataFileDateDisplay converts "YYYY-MM-DD" to "Mon, Jan 02" for console/PDF output.
func dataFileDateDisplay(date string) string {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return date
	}
	return t.Format("Mon, Jan 02")
}
