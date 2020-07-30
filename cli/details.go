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
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
)

type detail struct {
	statusText      *text.Text
	dangerText      *text.Text
	sparklineDanger *sparkline.SparkLine
	dangerSeries    *signalutils.Timeseries

	groupButton *button.Button
	sysText     *text.Text

	sparkline1   *sparkline.SparkLine
	sparkSeries1 *signalutils.Timeseries

	sparkline2   *sparkline.SparkLine
	sparkSeries2 *signalutils.Timeseries

	sparkline3   *sparkline.SparkLine
	sparkSeries3 *signalutils.Timeseries

	bottleneckText *text.Text
	riskText       *text.Text

	pausedShow bool

	group string
	rc    container.Option
}

func newDetails(group string, opt Option, ps *perfstat.Perfstat) (*detail, error) {

	h := &detail{group: group}

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
	ts := signalutils.NewTimeseries(4 * time.Minute)
	h.dangerSeries = &ts

	//HEADER
	h.groupButton, err = createButton(h.group, cell.ColorYellow)
	if err != nil {
		return nil, err
	}
	h.sparkline1, err = sparkline.New(sparkline.Color(cell.ColorYellow))
	if err != nil {
		return nil, err
	}
	ts1 := signalutils.NewTimeseries(4 * time.Minute)
	h.sparkSeries1 = &ts1

	h.sparkline2, err = sparkline.New(sparkline.Color(cell.ColorYellow))
	if err != nil {
		return nil, err
	}
	ts2 := signalutils.NewTimeseries(4 * time.Minute)
	h.sparkSeries2 = &ts2

	h.sparkline3, err = sparkline.New(sparkline.Color(cell.ColorYellow))
	if err != nil {
		return nil, err
	}
	ts3 := signalutils.NewTimeseries(4 * time.Minute)
	h.sparkSeries3 = &ts3

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
									container.PlaceWidget(h.groupButton),
								),
								container.Right(
									container.PlaceWidget(h.sparkline3),
									container.PaddingTop(0),
									container.PaddingBottom(2),
									container.PaddingLeft(3),
									container.PaddingRight(3),
								),
							),
						),
						container.Right(
							container.SplitVertical(
								container.Left(
									container.PlaceWidget(h.sparkline1),
									container.PaddingTop(0),
									container.PaddingBottom(2),
									container.PaddingLeft(3),
									container.PaddingRight(3),
								),
								container.Right(
									container.PlaceWidget(h.sparkline2),
									container.PaddingTop(0),
									container.PaddingBottom(2),
									container.PaddingLeft(3),
									container.PaddingRight(3),
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

	h.rc = c

	return h, nil
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
	sparklineDanger2, err := addSparkline(perc(od), h.dangerSeries, "", true)
	if err != nil {
		return err
	}
	*h.sparklineDanger = *sparklineDanger2

	//HEADER
	scc := ps.Score("", fmt.Sprintf("%s.*", h.group))

	color := cell.ColorRed
	bvalue := perc(scc)
	if bvalue < 80 {
		color = cell.ColorYellow
	}
	if bvalue < 5 {
		color = cell.ColorGreen
	}
	groupButton2, err := button.New(fmt.Sprintf("[%d] %s", bvalue, strings.ToUpper(h.group)),
		func() error { return nil },
		button.Width(15),
		button.Height(5),
		button.FillColor(color),
		button.ShadowColor(cell.ColorBlack))
	if err != nil {
		return err
	}
	*h.groupButton = *groupButton2

	if h.group == "cpu" {
		updateSparkSeriesTimeLoad(&detectors.ActiveStats.CPUStats.Total.User, h.sparkSeries1, "User", h.sparkline1, false)
		updateSparkSeriesTimeLoad(&detectors.ActiveStats.CPUStats.Total.IOWait, h.sparkSeries2, "IOWait", h.sparkline2, false)
		updateSparkSeriesTimeLoad(&detectors.ActiveStats.CPUStats.Total.Idle, h.sparkSeries3, "Total", h.sparkline3, true)

	} else if h.group == "mem" {
		used, ok := detectors.ActiveStats.MemStats.Used.Last()
		if ok {
			swapUsed, ok := detectors.ActiveStats.MemStats.SwapUsed.Last()
			if ok {
				updateSparkSeriesAbsoluteMax(used.Value+swapUsed.Value, "B", h.sparkSeries3, "Total", h.sparkline3, float64(detectors.ActiveStats.MemStats.Total+detectors.ActiveStats.MemStats.SwapTotal))
				updateSparkSeriesAbsoluteMax(used.Value, "B", h.sparkSeries1, "RAM", h.sparkline1, float64(detectors.ActiveStats.MemStats.Total))
				updateSparkSeriesAbsoluteMax(swapUsed.Value, "B", h.sparkSeries2, "SWAP", h.sparkline2, float64(detectors.ActiveStats.MemStats.SwapTotal))
			}
		}

	} else if h.group == "disk" {
		worstTotal := 0.0
		worstUsed := 0.0
		worstPerc := 0.0
		for _, part := range detectors.ActiveStats.DiskStats.Partitions {
			free, ok := part.Free.Last()
			if !ok {
				continue
			}
			used := float64(part.Total) - free.Value
			perc := used / float64(part.Total)
			if perc > worstPerc {
				worstTotal = float64(part.Total)
				worstUsed = used
				worstPerc = worstUsed / worstTotal
			}
		}
		updateSparkSeriesAbsoluteMax(worstUsed, "B", h.sparkSeries3, "Total", h.sparkline3, worstTotal)

		rwc := 0.0
		rwb := 0.0
		for _, disk := range detectors.ActiveStats.DiskStats.Disks {
			rc, ok := disk.ReadCount.Rate(4 * time.Second)
			if !ok {
				continue
			}
			wc, ok := disk.WriteCount.Rate(4 * time.Second)
			if !ok {
				continue
			}
			rwc = rwc + rc + wc

			rb, ok := disk.ReadBytes.Rate(4 * time.Second)
			if !ok {
				continue
			}
			wb, ok := disk.WriteBytes.Rate(4 * time.Second)
			if !ok {
				continue
			}
			rwb = rwb + rb + wb
		}

		updateSparkSeriesAbsoluteMax(rwc, "ops", h.sparkSeries1, "R/W", h.sparkline1, -1)
		updateSparkSeriesAbsoluteMax(rwb, "bps", h.sparkSeries2, "R/W", h.sparkline2, -1)

	} else if h.group == "net" {
		errorsTotal := 0.0
		for _, nic := range detectors.ActiveStats.NetStats.NICs {
			ei, ok := nic.ErrIn.Rate(4 * time.Second)
			if !ok {
				continue
			}
			eo, ok := nic.ErrOut.Rate(4 * time.Second)
			if !ok {
				continue
			}

			errorsTotal = errorsTotal + ei + eo
		}
		updateSparkSeriesAbsoluteMax(errorsTotal, "ops", h.sparkSeries3, "Errors", h.sparkline3, -1)

		rwc := 0.0
		rwb := 0.0
		for _, nic := range detectors.ActiveStats.NetStats.NICs {
			rc, ok := nic.PacketsRecv.Rate(4 * time.Second)
			if !ok {
				continue
			}
			wc, ok := nic.PacketsSent.Rate(4 * time.Second)
			if !ok {
				continue
			}
			rwc = rwc + rc + wc

			rb, ok := nic.BytesRecv.Rate(4 * time.Second)
			if !ok {
				continue
			}
			wb, ok := nic.BytesSent.Rate(4 * time.Second)
			if !ok {
				continue
			}
			rwb = rwb + rb + wb
		}

		updateSparkSeriesAbsoluteMax(rwc, "pps", h.sparkSeries1, "IN/OUT", h.sparkline1, -1)
		updateSparkSeriesAbsoluteMax(rwb, "bps", h.sparkSeries2, "IN/OUT", h.sparkline2, -1)

	}

	//BOTTLENECK DETAILS
	dr := ps.TopCriticity(0.01, "bottleneck", fmt.Sprintf("%s.*", h.group), false)
	h.bottleneckText.Write(detectionTxt(dr), text.WriteReplace())

	//RISK DETAILS
	dr = ps.TopCriticity(0.01, "risk", fmt.Sprintf("%s.*", h.group), false)
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

func updateSparkSeriesTimeLoad(timeLoadTs *signalutils.Timeseries, sts *signalutils.Timeseries, label string, sl *sparkline.SparkLine, inverse bool) {
	up1, ok := stats.TimeLoadPerc(timeLoadTs, 4*time.Second)
	if !ok {
		return
	}
	if inverse {
		up1 = 1 - up1
	}
	up := perc(up1)
	sparkline22, _ := addSparkline(up, sts, fmt.Sprintf("%s %d%%", label, up), true)
	*sl = *sparkline22
}

func updateSparkSeriesAbsoluteMax(value float64, unit string, sts *signalutils.Timeseries, label string, sl *sparkline.SparkLine, maxValue float64) {
	if maxValue != -1 {
		up := perc(value / maxValue)
		valuestr, unit2 := formatValueUnit(value, unit)
		sparkline22, _ := addSparkline(up, sts, fmt.Sprintf("%s %s%s %d%%", label, valuestr, unit2, up), true)
		*sl = *sparkline22
	} else {
		valuestr, unit2 := formatValueUnit(value, unit)
		up := int(value * 100)
		sparkline22, _ := addSparkline(up, sts, fmt.Sprintf("%s %s%s", label, valuestr, unit2), false)
		*sl = *sparkline22
	}
}

func (h *detail) rootContainer() container.Option {
	return h.rc
}
