package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/flaviostutz/perfstat"
	"github.com/flaviostutz/perfstat/detectors"
	"github.com/flaviostutz/perfstat/stats"
	"github.com/flaviostutz/signalutils"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
)

type detail struct {
	statusText      *text.Text
	dangerText      *text.Text
	sparklineDanger *sparkline.SparkLine
	dangerSeries    *signalutils.Timeseries

	sysButton *button.Button
	sysText   *text.Text

	sparkline1   *sparkline.SparkLine
	sparkSeries1 *signalutils.Timeseries

	sparkline2   *sparkline.SparkLine
	sparkSeries2 *signalutils.Timeseries

	sparkline3   *sparkline.SparkLine
	sparkSeries3 *signalutils.Timeseries

	donut1  *donut.Donut
	header2 interface{}

	bottleneckText *text.Text
	riskText       *text.Text

	pausedShow bool

	sys string
}

func newDetail(sys string) detail {
	return detail{sys: sys}
}

func (h *detail) build(opt Option, ps *perfstat.Perfstat) (container.Option, error) {

	//STATUS LINE
	titleText, err := text.New()
	if err != nil {
		return nil, err
	}
	titleText.Write("Perfstat")

	h.statusText, err = text.New()
	if err != nil {
		return nil, err
	}

	h.dangerText, err = text.New()
	if err != nil {
		return nil, err
	}

	h.sparklineDanger, err = sparkline.New(
		sparkline.Color(cell.ColorYellow),
	)
	ts := signalutils.NewTimeseries(10 * time.Minute)
	h.dangerSeries = &ts

	//HEADER
	h.sysButton, err = createButton(h.sys, cell.ColorYellow)
	if err != nil {
		return nil, err
	}
	h.sparkline1, err = sparkline.New(sparkline.Color(cell.ColorYellow))
	if err != nil {
		return nil, err
	}
	ts1 := signalutils.NewTimeseries(10 * time.Minute)
	h.sparkSeries1 = &ts1

	h.sparkline2, err = sparkline.New(sparkline.Color(cell.ColorYellow))
	if err != nil {
		return nil, err
	}
	ts2 := signalutils.NewTimeseries(10 * time.Minute)
	h.sparkSeries2 = &ts2

	h.sparkline3, err = sparkline.New(sparkline.Color(cell.ColorYellow))
	if err != nil {
		return nil, err
	}
	ts3 := signalutils.NewTimeseries(10 * time.Minute)
	h.sparkSeries3 = &ts3

	h.donut1, err = donut.New(donut.CellOpts(cell.FgColor(cell.ColorYellow)))
	if err != nil {
		return nil, err
	}

	h.header2 = h.sparkline3
	if h.sys == "net" {
		h.header2 = h.donut1
	}

	//DETAIL TEXTS
	h.bottleneckText, err = text.New()
	if err != nil {
		return nil, err
	}
	h.riskText, err = text.New()
	if err != nil {
		return nil, err
	}

	//place components
	c := container.SplitHorizontal(

		//STATUS
		container.Top(
			container.SplitVertical(
				container.Left(
					container.SplitVertical(
						container.Left(
							container.PlaceWidget(titleText),
							container.PaddingLeft(1),
						),
						container.Right(
							container.PlaceWidget(h.statusText),
						),
					),
				),
				container.Right(
					container.SplitVertical(
						container.Left(
							container.PlaceWidget(h.dangerText),
						),
						container.Right(
							container.PlaceWidget(h.sparklineDanger),
							container.PaddingRight(1),
						),
					),
				),
			),
		),
		container.Bottom(
			container.SplitHorizontal(
				//HEADER
				container.Top(
					container.PaddingTop(1),
					container.SplitVertical(
						container.Left(
							container.SplitVertical(
								container.Left(
									container.PlaceWidget(h.sysButton),
								),
								container.Right(
									container.PlaceWidget(h.header2.(widgetapi.Widget)),
								),
							),
						),
						container.Right(
							container.SplitVertical(
								container.Left(
									container.PlaceWidget(h.sparkline1),
									container.PaddingBottom(1),
									container.PaddingTop(1),
									container.PaddingLeft(1),
									container.PaddingRight(1),
								),
								container.Right(
									container.PlaceWidget(h.sparkline2),
									container.PaddingBottom(1),
									container.PaddingTop(1),
									container.PaddingLeft(1),
									container.PaddingRight(1),
								),
							),
						),
					),
				),

				//DETAILS TEXT
				container.Bottom(
					container.SplitVertical(
						container.Left(
							container.PaddingTop(1),
							container.PaddingLeft(1),
							container.BorderTitle("BOTTLENECKS"),
							container.Border(linestyle.Light),
							container.Border(linestyle.Round),
							container.MarginRight(1),
							container.PlaceWidget(h.bottleneckText),
						),
						container.Right(
							container.PaddingTop(1),
							container.PaddingLeft(1),
							container.PaddingRight(1),
							container.BorderTitle("RISKS"),
							container.Border(linestyle.Round),
							container.PlaceWidget(h.riskText),
						),
					),
				),
				container.SplitFixed(7),
			),
		),
		container.SplitFixed(1),
	)

	return c, nil
}

func (h *detail) update(opt Option, ps *perfstat.Perfstat, paused bool) error {

	//STATUS
	if paused {
		status := " "
		if h.pausedShow {
			status = "PAUSED"
		}
		h.statusText.Write(status, text.WriteReplace())
		h.pausedShow = !h.pausedShow
		return nil
	}
	h.statusText.Write("RUNNING", text.WriteReplace())

	//DANGER LEVEL
	danger := dangerLevel(ps)
	h.dangerText.Write(fmt.Sprintf("Danger: %d", danger), text.WriteReplace())

	//DANGER GRAPH
	od := ps.Score("", "")
	sparklineDanger2, err := addSparkline(perc(od), h.dangerSeries, "")
	if err != nil {
		return err
	}
	*h.sparklineDanger = *sparklineDanger2

	//HEADER
	scc := ps.Score("", fmt.Sprintf("%s.*", h.sys))

	color := cell.ColorRed
	bvalue := perc(scc)
	if bvalue < 80 {
		color = cell.ColorYellow
	}
	if bvalue < 5 {
		color = cell.ColorGreen
	}
	sysButton2, err := button.New(fmt.Sprintf("[%d] %s", bvalue, strings.ToUpper(h.sys)),
		func() error { return nil },
		button.Width(15),
		button.Height(5),
		button.FillColor(color),
		button.ShadowColor(cell.ColorBlack))
	if err != nil {
		return err
	}
	*h.sysButton = *sysButton2

	if h.sys == "cpu" {
		//USER TIME
		up1, ok := stats.TimeLoadPerc(&detectors.ActiveStats.CPUStats.Total.User, 4*time.Second)
		if !ok {
			up1 = -1
		}
		up := perc(up1)
		sparkline12, err := addSparkline(up, h.sparkSeries1, fmt.Sprintf("User %d%%", up))
		if err != nil {
			return err
		}
		*h.sparkline1 = *sparkline12

		//IOWAIT TIME
		up1, ok = stats.TimeLoadPerc(&detectors.ActiveStats.CPUStats.Total.IOWait, 4*time.Second)
		if !ok {
			up1 = -1
		}
		up = perc(up1)
		sparkline22, err := addSparkline(up, h.sparkSeries2, fmt.Sprintf("IOWait %d%%", up))
		if err != nil {
			return err
		}
		*h.sparkline2 = *sparkline22

		//OVERALL LOAD
		up1, ok = stats.TimeLoadPerc(&detectors.ActiveStats.CPUStats.Total.Idle, 4*time.Second)
		if !ok {
			up1 = -1
		}
		up = perc(1 - up1)
		dc := cell.ColorRed
		if up < 80 {
			dc = cell.ColorYellow
		}
		if up < 20 {
			dc = cell.ColorGreen
		}
		donut12, err := donut.New(donut.CellOpts(cell.FgColor(dc)))
		if err != nil {
			return err
		}
		donut12.Percent(up)
		*h.donut1 = *donut12

	}

	//BOTTLENECK DETAILS
	dr := ps.TopCriticity(0.01, "bottleneck", fmt.Sprintf("%s.*", h.sys), false)
	h.bottleneckText.Write(detectionTxt(dr), text.WriteReplace())

	//RISK DETAILS
	dr = ps.TopCriticity(0.01, "risk", fmt.Sprintf("%s.*", h.sys), false)
	h.riskText.Write(detectionTxt(dr), text.WriteReplace())

	return nil
}

func (h *detail) onEvent(evt *terminalapi.Keyboard) {
}

func createButton(sys string, color cell.Color) (*button.Button, error) {
	b, err := button.New(fmt.Sprintf("[%d] %s", 0, strings.ToUpper(sys)),
		func() error { return nil },
		button.Width(15),
		button.Height(5),
		button.FillColor(color),
		button.ShadowColor(cell.ColorBlack))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func detectionTxt(dr []detectors.DetectionResult) string {
	r := ""
	for _, d := range dr {
		if r == "" {
			r = renderDR(d)
		} else {
			r = fmt.Sprintf("%s\n%s", r, renderDR(d))
		}
		for _, rel := range d.Related {
			pn, pv, unit := formatResPropertyValue(rel)
			r = fmt.Sprintf("%s\n  > %s %s %s%s", r, rel.Name, pn, pv, unit)
		}
	}
	return r
}
