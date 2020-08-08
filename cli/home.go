package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/flaviostutz/perfstat"
	"github.com/flaviostutz/signalutils"
	"github.com/jedib0t/go-pretty/table"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
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

	rc container.Option
}

func newHome(opt Option, ps *perfstat.Perfstat) (*home, error) {

	h := &home{}

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

	h.cpuButton, h.cpuText, err = subsystemBox(nil, nil, "CPU", 0, "1", 15, 5, " ")
	if err != nil {
		return nil, err
	}

	h.memButton, h.memText, err = subsystemBox(nil, nil, "MEM", 0, "2", 15, 5, " ")
	if err != nil {
		return nil, err
	}

	h.diskButton, h.diskText, err = subsystemBox(nil, nil, "DISK", 0, "3", 15, 5, " ")
	if err != nil {
		return nil, err
	}

	h.netButton, h.netText, err = subsystemBox(nil, nil, "NET", 0, "4", 15, 5, " ")
	if err != nil {
		return nil, err
	}

	h.diskButtonr, h.diskTextr, err = subsystemBox(nil, nil, "DISK", 0, "3", 15, 3, " ")
	if err != nil {
		return nil, err
	}

	h.memButtonr, h.memTextr, err = subsystemBox(nil, nil, "MEM", 0, "2", 15, 3, " ")
	if err != nil {
		return nil, err
	}

	h.netButtonr, h.netTextr, err = subsystemBox(nil, nil, "NET", 0, "4", 15, 3, " ")
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
											container.ID("cpuButton"),
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
											container.ID("memButton"),
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
											container.ID("diskButton"),
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
											container.ID("netButton"),
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
													container.ID("diskButtonr"),
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
													container.ID("memButtonr"),
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
											container.ID("netButtonr"),
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

	h.rc = c

	return h, nil
}

func (h *home) update(opt Option, ps *perfstat.Perfstat, paused bool, term *termbox.Terminal) error {

	tw := term.Size().X
	th := term.Size().Y

	bw := int(math.Min(math.Max(float64(tw/7), 6.0), 18.0))
	bh := int(math.Min(math.Max(float64(th)/6.0, 1.0), 5.0))
	bh2 := int(math.Max(float64(bh-2), 1))

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
	_, err := resolveSparkline(h.sparklineDanger, perc(od), h.dangerSeries, "", true)
	if err != nil {
		return err
	}
	danger := dangerLevel(ps)
	h.dangerText.Write(fmt.Sprintf("Danger: %d", danger), text.WriteReplace())

	//BOTTLENECK
	scc := ps.Score("bottleneck", "cpu.*")
	drc := ps.TopCriticity(0.01, "bottleneck", "cpu.*", false)
	cpuButton2, _, err := subsystemBox(h.cpuButton, h.cpuText, "CPU", int(math.Round(scc*100.0)), "2", bw, bh, renderDetectionResults(drc))
	if err != nil {
		return err
	}
	rootc.Update("cpuButton", container.PlaceWidget(cpuButton2))

	scm := ps.Score("bottleneck", "mem.*")
	drm := ps.TopCriticity(0.01, "bottleneck", "mem.*", false)
	memButton2, _, err := subsystemBox(h.cpuButton, h.memText, "MEM", int(math.Round(scm*100.0)), "3", bw, bh, renderDetectionResults(drm))
	if err != nil {
		return err
	}
	rootc.Update("memButton", container.PlaceWidget(memButton2))

	scd := ps.Score("bottleneck", "disk.*")
	drd := ps.TopCriticity(0.01, "bottleneck", "disk.*", false)
	diskButton2, _, err := subsystemBox(h.cpuButton, h.diskText, "DISK", int(math.Round(scd*100.0)), "4", bw, bh, renderDetectionResults(drd))
	if err != nil {
		return err
	}
	rootc.Update("diskButton", container.PlaceWidget(diskButton2))

	scn := ps.Score("bottleneck", "net.*")
	drn := ps.TopCriticity(0.01, "bottleneck", "net.*", false)
	netButton2, _, err := subsystemBox(h.netButton, h.netText, "NET", int(math.Round(scn*100.0)), "5", bw, bh, renderDetectionResults(drn))
	if err != nil {
		return err
	}
	rootc.Update("netButton", container.PlaceWidget(netButton2))

	//RISKS
	scd = ps.Score("risk", "disk.*")
	drd = ps.TopCriticity(0.01, "risk", "disk.*", false)
	diskButton2r, _, err := subsystemBox(h.diskButtonr, h.diskTextr, "DISK", int(math.Round(scd*100.0)), "4", bw, bh2, renderDetectionResults(drd))
	if err != nil {
		return err
	}
	rootc.Update("diskButtonr", container.PlaceWidget(diskButton2r))

	scm = ps.Score("risk", "mem.*")
	drm = ps.TopCriticity(0.01, "risk", "mem.*", false)
	memButton2r, _, err := subsystemBox(h.memButtonr, h.memTextr, "MEM", int(math.Round(scm*100.0)), "3", bw, bh2, renderDetectionResults(drm))
	if err != nil {
		return err
	}
	rootc.Update("memButtonr", container.PlaceWidget(memButton2r))

	scn = ps.Score("risk", "net.*")
	drn = ps.TopCriticity(0.01, "risk", "net.*", false)
	netButton2r, _, err := subsystemBox(h.netButtonr, h.netTextr, "NET", int(math.Round(scn*100.0)), "5", bw, bh2, renderDetectionResults(drn))
	if err != nil {
		return err
	}
	rootc.Update("netButtonr", container.PlaceWidget(netButton2r))

	//RELATED
	t := table.NewWriter()
	dr := ps.TopCriticity(0.01, "", "", false)
	related := make(map[string]string, 0)
	for _, d0 := range dr {
		for _, r00 := range d0.Related {
			k := fmt.Sprintf("%s-%s", r00.Typ, r00.Name)
			_, found := related[k]
			if !found {
				pn, pv, unit := formatResPropertyValue(r00)
				related[k] = "OK"
				t.AppendRows([]table.Row{
					{fmt.Sprintf("%.0f", math.Round(d0.Score*100)), r00.Name, pn, fmt.Sprintf("%s%s", pv, unit), d0.ID, d0.Typ},
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

func subsystemBox(btn *button.Button, tx *text.Text, blabel string, bvalue int, bKeyText string, bwidth int, bheight int, status string) (*button.Button, *text.Text, error) {
	//button
	color := cell.ColorRed
	if bvalue < 80 {
		color = cell.ColorYellow
	}
	if bvalue < 5 {
		color = cell.ColorGreen
	}

	var err error
	label := fmt.Sprintf("[%d] %s (%s)", bvalue, blabel, bKeyText)
	// if btn == nil {
	btn, err = button.New(label,
		func() error {
			showScreen(strings.ToLower(blabel))
			return nil
		},
		button.Width(bwidth),
		button.Height(bheight),
		button.FillColor(color),
		button.ShadowColor(cell.ColorBlack))
	if err != nil {
		return nil, nil, err
	}
	// }

	//status text
	if tx == nil {
		tx, err = text.New()
		if err != nil {
			return nil, nil, err
		}
	}
	tx.Write(fmt.Sprintf("%s", status), text.WriteReplace())

	return btn, tx, nil
}

func (h *home) rootContainer() container.Option {
	return h.rc
}
