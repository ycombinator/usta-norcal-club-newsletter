package formatters

import (
	"io"
	"os"

	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
)

type PDFFormatter struct {
	reader io.Reader
	writer io.Writer
}

func NewPDFFormatter() *PDFFormatter {
	return &PDFFormatter{reader: os.Stdin, writer: os.Stdout}
}

func (p *PDFFormatter) Format(n *core.Newsletter, cfg Config) error {
	data, err := Prepare(n, cfg, p.reader, p.writer)
	if err != nil {
		return err
	}

	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	cellTextProps := props.Text{Size: 8, Top: 2}

	if len(data.PastMatches) > 0 {
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
			date, first, outcome, locOpponent := formatAnnotatedMatch(am, data.Org, data.OrgNames, p.reader, p.writer)
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
	}

	if len(data.FutureMatches) > 0 {
		m.Row(10, func() {})

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
			date, first, _, locOpponent := formatFutureMatch(match, data.Org, data.OrgNames, p.reader, p.writer)
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
	}

	if err := data.Save(); err != nil {
		return err
	}

	path, err := OutputPath(cfg.OutputDir, OutputFilename(data.Org.ShortName(), "newsletter", "pdf"))
	if err != nil {
		return err
	}
	return m.OutputFileAndClose(path)
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
