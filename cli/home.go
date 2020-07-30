package main

import (
	"fmt"
	"math"
	"time"

	"github.com/flaviostutz/perfstat"
	"github.com/flaviostutz/perfstat/detectors"
	"github.com/flaviostutz/signalutils"
	"github.com/jedib0t/go-pretty/table"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
)

type home struct {
	statusText      *text.Text
	dangerText      *text.Text
	sparklineDanger *sparkline.SparkLine

	cpuButton  *button.Button
	cpuText    *text.Text
	memButton  *button.Button
	memText    *text.Text
	diskButton *button.Button
	diskText   *text.Text
	netButton  *button.Button
	netText    *text.Text

	memButtonr  *button.Button
	memTextr    *text.Text
	diskButtonr *button.Button
	diskTextr   *text.Text
	netButtonr  *button.Button
	netTextr    *text.Text

	relatedText  *text.Text
	dangerSeries *signalutils.Timeseries
	pausedShow   bool
}

func (h *home) build(opt Option, ps *perfstat.Perfstat) (container.Option, error) {

	//prepare widgets
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

	h.cpuButton, h.cpuText, err = subsystemBox("CPU", 0, 49, "1", 15, 5, "")
	if err != nil {
		return nil, err
	}

	h.memButton, h.memText, err = subsystemBox("MEM", 0, 50, "2", 15, 5, "")
	if err != nil {
		return nil, err
	}

	h.diskButton, h.diskText, err = subsystemBox("DISK", 0, 51, "3", 15, 5, "")
	if err != nil {
		return nil, err
	}

	h.netButton, h.netText, err = subsystemBox("NET", 0, 52, "4", 15, 5, "")
	if err != nil {
		return nil, err
	}

	h.diskButtonr, h.diskTextr, err = subsystemBox("DISK", 0, 51, "3", 15, 3, "")
	if err != nil {
		return nil, err
	}

	h.memButtonr, h.memTextr, err = subsystemBox("MEM", 0, 50, "2", 15, 3, "")
	if err != nil {
		return nil, err
	}

	h.netButtonr, h.netTextr, err = subsystemBox("NET", 0, 52, "4", 15, 3, "")
	if err != nil {
		return nil, err
	}

	h.relatedText, err = text.New()
	if err != nil {
		return nil, err
	}

	//place components
	c := container.SplitHorizontal(
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
				container.Top(
					container.BorderTitle("BOTTLENECKS"),
					container.Border(linestyle.Light),
					container.PaddingTop(1),
					container.Border(linestyle.Round),
					container.SplitHorizontal(
						container.Top(
							container.SplitVertical(
								container.Left(
									container.SplitVertical(
										container.Left(
											container.PlaceWidget(h.cpuButton),
										),
										container.Right(
											container.PlaceWidget(h.cpuText),
										),
										container.SplitFixed(20),
									),
								),
								container.Right(
									container.SplitVertical(
										container.Left(
											container.PlaceWidget(h.memButton),
										),
										container.Right(
											container.PlaceWidget(h.memText),
										),
										container.SplitFixed(20),
									),
								),
							),
						),
						container.Bottom(
							container.SplitVertical(
								container.Left(
									container.SplitVertical(
										container.Left(
											container.PlaceWidget(h.diskButton),
										),
										container.Right(
											container.PlaceWidget(h.diskText),
										),
										container.SplitFixed(20),
									),
								),
								container.Right(
									container.SplitVertical(
										container.Left(
											container.PlaceWidget(h.netButton),
										),
										container.Right(
											container.PlaceWidget(h.netText),
										),
										container.SplitFixed(20),
									),
								),
							),
						),
					),
				),

				container.Bottom(
					container.SplitVertical(
						container.Left(
							container.PaddingTop(1),
							container.BorderTitle("RISKS"),
							container.Border(linestyle.Light),
							container.Border(linestyle.Round),
							container.MarginRight(1),
							container.SplitHorizontal(
								container.Top(
									container.SplitHorizontal(
										container.Top(
											container.SplitVertical(
												container.Left(
													container.PlaceWidget(h.diskButtonr),
												),
												container.Right(
													container.PlaceWidget(h.diskTextr),
												),
												container.SplitFixed(20),
											),
										),
										container.Bottom(
											container.SplitVertical(
												container.Left(
													container.PlaceWidget(h.memButtonr),
												),
												container.Right(
													container.PlaceWidget(h.memTextr),
												),
												container.SplitFixed(20),
											),
										),
									),
								),
								container.Bottom(
									container.SplitVertical(
										container.Left(
											container.PlaceWidget(h.netButtonr),
										),
										container.Right(
											container.PlaceWidget(h.netTextr),
										),
										container.SplitFixed(20),
									),
								),
								container.SplitPercent(70),
							),
						),
						container.Right(
							container.PaddingTop(1),
							container.PaddingLeft(1),
							container.PaddingRight(1),
							container.BorderTitle("Top Related"),
							container.Border(linestyle.Round),
							container.PlaceWidget(h.relatedText),
						),
						container.SplitPercent(51),
					),
				),
				container.SplitPercent(55),
			),
		),
		container.SplitFixed(1),
	)

	return c, nil
}

func (h *home) update(opt Option, ps *perfstat.Perfstat, paused bool) error {

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
	od := ps.Score("", "")
	sparklineDanger2, err := addSparkline(perc(od), h.dangerSeries, "", true)
	if err != nil {
		return err
	}
	danger := dangerLevel(ps)
	h.dangerText.Write(fmt.Sprintf("Danger: %d", danger), text.WriteReplace())
	*h.sparklineDanger = *sparklineDanger2

	//BOTTLENECK
	scc := ps.Score("bottleneck", "cpu.*")
	drc := ps.TopCriticity(0.01, "bottleneck", "cpu.*", false)
	cpuButton2, cpuText2, err := subsystemBox("CPU", int(math.Round(scc*100.0)), 49, "1", 15, 5, renderDetectionResults(drc))
	if err != nil {
		return err
	}
	*h.cpuButton = *cpuButton2
	*h.cpuText = *cpuText2

	scm := ps.Score("bottleneck", "mem.*")
	drm := ps.TopCriticity(0.01, "bottleneck", "mem.*", false)
	memButton2, memText2, err := subsystemBox("MEM", int(math.Round(scm*100.0)), 50, "2", 15, 5, renderDetectionResults(drm))
	if err != nil {
		return err
	}
	*h.memButton = *memButton2
	*h.memText = *memText2

	scd := ps.Score("bottleneck", "disk.*")
	drd := ps.TopCriticity(0.01, "bottleneck", "disk.*", false)
	diskButton2, diskText2, err := subsystemBox("DISK", int(math.Round(scd*100.0)), 51, "3", 15, 5, renderDetectionResults(drd))
	if err != nil {
		return err
	}
	*h.diskButton = *diskButton2
	*h.diskText = *diskText2

	scn := ps.Score("bottleneck", "net.*")
	drn := ps.TopCriticity(0.01, "bottleneck", "net.*", false)
	netButton2, netText2, err := subsystemBox("NET", int(math.Round(scn*100.0)), 52, "4", 15, 5, renderDetectionResults(drn))
	if err != nil {
		return err
	}
	*h.netButton = *netButton2
	*h.netText = *netText2

	//RISKS
	scd = ps.Score("risk", "disk.*")
	drd = ps.TopCriticity(0.01, "risk", "disk.*", false)
	diskButton2r, diskText2r, err := subsystemBox("DISK", int(math.Round(scd*100.0)), 51, "3", 15, 3, renderDetectionResults(drd))
	if err != nil {
		return err
	}
	*h.diskButtonr = *diskButton2r
	*h.diskTextr = *diskText2r

	scm = ps.Score("risk", "mem.*")
	drm = ps.TopCriticity(0.01, "risk", "mem.*", false)
	memButton2r, memText2r, err := subsystemBox("MEM", int(math.Round(scm*100.0)), 52, "4", 15, 3, renderDetectionResults(drm))
	if err != nil {
		return err
	}
	*h.memButtonr = *memButton2r
	*h.memTextr = *memText2r

	scn = ps.Score("risk", "net.*")
	drn = ps.TopCriticity(0.01, "risk", "net.*", false)
	netButton2r, netText2r, err := subsystemBox("NET", int(math.Round(scn*100.0)), 52, "4", 15, 3, renderDetectionResults(drn))
	if err != nil {
		return err
	}
	*h.netButtonr = *netButton2r
	*h.netTextr = *netText2r

	//RELATED
	t := table.NewWriter()
	dr := ps.TopCriticity(0.01, "", "", false)
	related := make([]detectors.Resource, 0)
	for _, d0 := range dr {
		for _, r00 := range d0.Related {
			found := false
			for _, r0 := range related {
				if r0 == r00 {
					found = true
				}
			}
			if !found {
				pn, pv, unit := formatResPropertyValue(r00)
				related = append(related, r00)
				t.AppendRows([]table.Row{
					{fmt.Sprintf("%.0f", math.Round(d0.Score*100)), r00.Name, pn, fmt.Sprintf("%s%s", pv, unit)},
				})
			}
		}
	}
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateFooter = false
	t.Style().Options.SeparateHeader = false
	t.Style().Options.SeparateRows = false
	t.Style().Options.DrawBorder = false
	tr := t.Render()
	h.relatedText.Write(fmt.Sprintf("%s ", tr), text.WriteReplace())

	return nil
}

func (h *home) onEvent(evt *terminalapi.Keyboard) {
}

func subsystemBox(blabel string, bvalue int, bKey keyboard.Key, bKeyText string, bwidth int, bheight int, status string) (*button.Button, *text.Text, error) {
	//button
	color := cell.ColorRed
	if bvalue < 80 {
		color = cell.ColorYellow
	}
	if bvalue < 5 {
		color = cell.ColorGreen
	}
	c1, err := button.New(fmt.Sprintf("[%d] %s (%s)", bvalue, blabel, bKeyText),
		func() error {
			return nil
		},
		button.GlobalKey(bKey),
		button.Width(bwidth),
		button.Height(bheight),
		button.FillColor(color),
		button.ShadowColor(cell.ColorBlack))
	if err != nil {
		return nil, nil, err
	}

	//status text
	c2, err := text.New()
	if err != nil {
		return nil, nil, err
	}
	c2.Write(status, text.WriteReplace())

	return c1, c2, nil
}
