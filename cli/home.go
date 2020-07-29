babasibpackage main

import (
	"fmt"
	"math"

	"github.com/flaviostutz/perfstat"
	"github.com/jedib0t/go-pretty/table"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
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

	relatedText *text.Text
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

	h.diskButtonr, h.diskTextr, err = subsystemBox("DISK", 0, 53, "5", 15, 3, "")
	if err != nil {
		return nil, err
	}

	h.memButtonr, h.memTextr, err = subsystemBox("MEM", 0, 54, "6", 15, 3, "")
	if err != nil {
		return nil, err
	}

	h.netButtonr, h.netTextr, err = subsystemBox("NET", 0, 55, "7", 15, 3, "")
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
					container.BorderTitle("Bottleneck"),
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
							container.BorderTitle("Risk"),
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

func (h *home) update(opt Option, ps *perfstat.Perfstat) error {

	//DANGER LEVEL
	ds := ps.TopCriticity(0, "", "", false)
	danger := 0.0
	for _, d := range ds {
		danger = danger + d.Score
	}
	h.dangerText.Write(fmt.Sprintf("Danger: %d", int(math.Round(danger*100))), text.WriteReplace())

	h.sparklineDanger.Add([]int{21, 23, 43, 47, 42, 20, 21, 23, 43, 47, 42, 20, 21, 23, 43, 47, 42, 20, 7})

	t := table.NewWriter()
	t.AppendRows([]table.Row{
		{1, "Arya", "Stark", 3000},
		{20, "Jon", "Snow", 2000, "You know nothing, Jon Snow!"},
	})
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateFooter = false
	t.Style().Options.SeparateHeader = false
	t.Style().Options.SeparateRows = false
	t.Style().Options.DrawBorder = false
	tr := t.Render()
	h.relatedText.Write(tr, text.WriteReplace())

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

	return nil
}

func (h *home) onEvent(evt *terminalapi.Keyboard) {
}
