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
	if len(data.PastMatches) == 0 {
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

	for i, am := range data.PastMatches {
		setRowColor(i, m)
		date, first, outcome, locOpponent := formatAnnotatedMatch(am, data.Org, data.OrgNames, cfg.Reader, cfg.Writer)
		m.Row(8, func() {
			m.Col(2, func() {
				m.Text(" "+date, cellTextProps)
			})
			m.Col(4, func() {
				m.Text(first, cellTextProps)
			})
			m.Col(1, func() {
				m.Text(outcome, cellTextProps)
			})
			m.Col(5, func() {
				m.Text(locOpponent, cellTextProps)
			})
		})
	}

	path, err := OutputPath(cfg.OutputDir, OutputFilename(data.Org.ShortName(), "recent", "pdf"))
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
	if len(data.FutureMatches) == 0 {
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

	for i, match := range data.FutureMatches {
		setRowColor(i, m)
		date, first, _, locOpponent := formatFutureMatch(match, data.Org, data.OrgNames, cfg.Reader, cfg.Writer)
		m.Row(8, func() {
			m.Col(3, func() {
				m.Text(" "+date, cellTextProps)
			})
			m.Col(4, func() {
				m.Text(first, cellTextProps)
			})
			m.Col(5, func() {
				m.Text(locOpponent, cellTextProps)
			})
		})
	}

	path, err := OutputPath(cfg.OutputDir, OutputFilename(data.Org.ShortName(), "upcoming", "pdf"))
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
