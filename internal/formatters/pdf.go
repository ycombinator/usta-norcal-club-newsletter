package formatters

import (
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
	"github.com/ycombinator/usta-norcal-club-newsletter/internal/core"
)

type PDFFormatter struct{}

func NewPDFFormatter() *PDFFormatter {
	return new(PDFFormatter)
}

func (p *PDFFormatter) Format(n *core.Newsletter, cfg Config) error {
	org := n.Organization()
	pastMatches, futureMatches := org.Matches(cfg.PastDuration, cfg.FutureDuration)

	m := pdf.NewMaroto(consts.Portrait, consts.A4)

	if len(pastMatches) > 0 {
		m.Row(10, func() {
			m.Col(12, func() {
				m.Text("Recent Matches", props.Text{
					Top:   3,
					Style: consts.Bold,
					Align: consts.Center,
				})
			})
		})

		for i, match := range pastMatches {
			setRowColor(i, m)
			date, first, outcome, locator, second := match.ForOrganization(org)
			m.Row(10, func() {
				m.Col(2, func() {
					m.Text(date.Format(" Mon, Jan 02"))
				})
				m.Col(4, func() {
					m.Text(first)
				})
				m.Col(2, func() {
					m.Text(outcome)
				})
				m.Col(4, func() {
					m.Text(locator + " " + second)
				})
			})
		}
	}

	if len(futureMatches) > 0 {
		m.Line(10)
		
		m.Row(10, func() {
			m.Col(12, func() {
				m.Text("Future	 Matches", props.Text{
					Top:   3,
					Style: consts.Bold,
					Align: consts.Center,
				})
			})
		})

		for i, match := range futureMatches {
			setRowColor(i, m)
			date, first, _, locator, second := match.ForOrganization(org)
			m.Row(10, func() {
				m.Col(2, func() {
					m.Text(date.Format(" Mon, Jan 02"))
				})
				m.Col(5, func() {
					m.Text(first)
				})
				m.Col(5, func() {
					m.Text(locator + " " + second)
				})
			})
		}
	}

	return m.OutputFileAndClose("./newsletter.pdf")
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
