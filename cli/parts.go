package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/flaviostutz/perfstat/detectors"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/text"
)

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
		func() error { return nil },
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

func renderDetectionResults(drs []detectors.DetectionResult) string {

	// t := table.NewWriter()
	// dr := ps.TopCriticity(0.01, "", "", true)
	// related := make([]detectors.Resource, 0)
	// for _, d0 := range dr {
	// 	found := false
	// 	for _, r0 := range related {
	// 		if r0 == d0.Res {
	// 			found = true
	// 		}
	// 	}
	// 	if !found {
	// 		pn, pv, unit := formatResPropertyValue(d0)
	// 		related = append(related, d0.Res)
	// 		t.AppendRows([]table.Row{
	// 			{fmt.Sprintf("%.0f", math.Round(d0.Score*100)), d0.Res.Name, pn, fmt.Sprintf("%s%s", pv, unit)},
	// 		})
	// 	}
	// }
	// t.Style().Options.SeparateColumns = false
	// t.Style().Options.SeparateFooter = false
	// t.Style().Options.SeparateHeader = false
	// t.Style().Options.SeparateRows = false
	// t.Style().Options.DrawBorder = false
	// tr := t.Render()

	res := ""
	for _, dr := range drs {
		if res == "" {
			res = renderDR(dr)
		} else {
			res = fmt.Sprintf("%s\n%s", res, renderDR(dr))
		}
	}
	return res
}

func renderDR(dr detectors.DetectionResult) string {
	idx := strings.Index(dr.ID, "-")
	if idx == -1 {
		return "ERROR"
	}
	idn := dr.ID[idx+1:]
	_, valueStr, unit := formatResPropertyValue(dr.Res)

	return fmt.Sprintf("%d %s %s=%s%s", perc(dr.Score), idn, dr.Res.Name, valueStr, unit)
}

func perc(v float64) int {
	return int(math.Round(v * 100))
}

func formatResPropertyValue(res detectors.Resource) (formattedName string, formattedValue string, unit string) {
	idx2 := strings.LastIndex(res.PropertyName, "-")
	if idx2 == -1 {
		return "ERR", "", ""
	}

	formattedName = res.PropertyName[:idx2]
	unit = res.PropertyName[idx2+1:]
	value := res.PropertyValue

	if unit == "bps" {
		unit = "Bps"
	} else if unit == "perc" {
		value = res.PropertyValue * 100
		unit = "%"
	}

	if value > 1000 {
		value = value / 1000
		unit = fmt.Sprintf("K%s", unit)
	} else if value > 1000000 {
		value = value / 1000000
		unit = fmt.Sprintf("M%s", unit)
	}

	formattedValue = fmt.Sprintf("%.2f", value)
	if value <= 100 {
		formattedValue = fmt.Sprintf("%.0f", value)
	}

	return formattedName, formattedValue, unit
}
