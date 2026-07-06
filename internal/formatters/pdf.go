package formatters

import (
	"fmt"

	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

type PDFFormatter struct{}

func NewPDFFormatter() *PDFFormatter {
	return &PDFFormatter{}
}

func (p *PDFFormatter) FormatRecent(data *PreparedData, cfg Config) error {
	if !data.hasPastMatches() {
		return nil
	}

	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	cellTextProps := props.Text{Size: 8, Top: 2}

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text("Recent Matches", props.Text{
				Top:   3,
				Style: consts.Bold,
				Align: consts.Center,
			})
		})
	})

	if data.DataFile != nil {
		for i, rec := range data.DataFile.PastMatches {
			i := i
			rec := rec
			setRowColor(i, m)
			date := dataFileDateDisplay(rec.Date)
			teamStr := data.DataFile.OrgShortName + " " + rec.GenderEmoji + rec.Level + rec.Superscript
			outcome := consoleOutcome(rec)
			locOpponent := consoleLocOpponent(rec.IsHome, rec.Opponent)
			m.Row(8, func() {
				m.Col(2, func() { m.Text(" "+date, cellTextProps) })
				m.Col(4, func() { m.Text(teamStr, cellTextProps) })
				m.Col(1, func() { m.Text(outcome, cellTextProps) })
				m.Col(5, func() { m.Text(locOpponent, cellTextProps) })
			})
		}
	} else {
		for i, am := range data.PastMatches {
			i := i
			am := am
			setRowColor(i, m)
			date, first, outcome, locOpponent := formatAnnotatedMatch(am, data.Org, data.OrgNames, cfg.Reader, cfg.Writer)
			m.Row(8, func() {
				m.Col(2, func() { m.Text(" "+date, cellTextProps) })
				m.Col(4, func() { m.Text(first, cellTextProps) })
				m.Col(1, func() { m.Text(outcome, cellTextProps) })
				m.Col(5, func() { m.Text(locOpponent, cellTextProps) })
			})
		}
	}

	path, err := OutputPath(cfg.OutputDir, OutputFilename(data.orgShortName(), "recent", "pdf"))
	if err != nil {
		return err
	}
	if err := m.OutputFileAndClose(path); err != nil {
		return err
	}
	fmt.Fprintln(cfg.Writer, "Wrote", path)
	return nil
}

func (p *PDFFormatter) FormatUpcoming(data *PreparedData, cfg Config) error {
	if !data.hasUpcomingMatches() {
		return nil
	}

	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	cellTextProps := props.Text{Size: 8, Top: 2}

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text("Upcoming Matches", props.Text{
				Top:   3,
				Style: consts.Bold,
				Align: consts.Center,
			})
		})
	})

	if data.DataFile != nil {
		for i, rec := range data.DataFile.FutureMatches {
			i := i
			rec := rec
			setRowColor(i, m)
			date := dataFileDateDisplay(rec.Date)
			if rec.Time != "" {
				date += " " + dataFileMatchTime(rec.Time)
			}
			teamStr := data.DataFile.OrgShortName + " " + rec.GenderEmoji + rec.Level + rec.Superscript
			locOpponent := consoleLocOpponent(rec.IsHome, rec.Opponent)
			if rec.LocationNote != "" {
				locOpponent += " (at " + rec.LocationNote + ")"
			}
			m.Row(8, func() {
				m.Col(3, func() { m.Text(" "+date, cellTextProps) })
				m.Col(4, func() { m.Text(teamStr, cellTextProps) })
				m.Col(5, func() { m.Text(locOpponent, cellTextProps) })
			})
		}
	} else {
		for i, match := range data.FutureMatches {
			i := i
			match := match
			setRowColor(i, m)
			date, first, _, locOpponent := formatFutureMatch(match, data.Org, data.OrgNames, cfg.Reader, cfg.Writer)
			m.Row(8, func() {
				m.Col(3, func() { m.Text(" "+date, cellTextProps) })
				m.Col(4, func() { m.Text(first, cellTextProps) })
				m.Col(5, func() { m.Text(locOpponent, cellTextProps) })
			})
		}
	}

	path, err := OutputPath(cfg.OutputDir, OutputFilename(data.orgShortName(), "upcoming", "pdf"))
	if err != nil {
		return err
	}
	if err := m.OutputFileAndClose(path); err != nil {
		return err
	}
	fmt.Fprintln(cfg.Writer, "Wrote", path)
	return nil
}

func setRowColor(rowIndex int, m pdf.Maroto) {
	lightGrayColor := color.Color{Red: 200, Green: 200, Blue: 200}
	whiteColor := color.NewWhite()

	if rowIndex%2 == 0 {
		m.SetBackgroundColor(lightGrayColor)
	} else {
		m.SetBackgroundColor(whiteColor)
	}
}
