package main

import (
	"context"

	"github.com/flaviostutz/perfstat"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/button"
)

type home struct {
}

func (h home) build(ctx context.Context, opt Option, ps *perfstat.Perfstat) ([]container.Option, error) {
	cpuBox, err := bottlRect("CPU")
	if err != nil {
		return nil, err
	}

	memBox, err := bottlRect("MEM")
	if err != nil {
		return nil, err
	}

	diskBox, err := bottlRect("DISK")
	if err != nil {
		return nil, err
	}

	netBox, err := bottlRect("DISK")
	if err != nil {
		return nil, err
	}

	builder := grid.New()
	builder.Add(
		grid.Widget(grid.RowHeightPerc(55), container.Border(linestyle.Light)),
		//BOTTLENECK ROW
		grid.RowHeightPerc(55,
			grid.RowHeightPerc(55,
				grid.ColWidthPerc(50,
					grid.ColWidthPerc(30,
						grid.Widget(
							cpuBox,
							container.Border(linestyle.None),
						),
					),
					grid.ColWidthPerc(70,
						grid.Widget(
							memBox,
							container.Border(linestyle.None),
						),
					),
					grid.ColWidthPerc(30,
						grid.Widget(
							cpuBox,
							container.Border(linestyle.None),
						),
					),
					grid.ColWidthPerc(70,
						grid.Widget(
							memBox,
							container.Border(linestyle.None),
						),
					),
				),
			),
			container.Border(linestyle.None),
		),
	)
	return builder.Build()
}

func (h home) onEvent(evt *terminalapi.Keyboard) {
}

func bottlRect(label string) (widgetapi.Widget, error) {
	return button.New(label, func() error { return nil })
}
