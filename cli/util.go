package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/flaviostutz/perfstat"
	"github.com/flaviostutz/perfstat/detectors"
	"github.com/flaviostutz/signalutils"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/sparkline"
)

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

	formattedValue, unit2 := formatValueUnit(value, unit)

	return formattedName, formattedValue, unit2
}

func formatValueUnit(value float64, unit string) (v string, u string) {
	// if true {
	// 	return fmt.Sprintf("%.2f", value), unit
	// }
	if unit == "bps" {
		unit = "Bps"
	} else if unit == "b" {
		unit = "B"
	} else if unit == "perc" {
		value = value * 100
		unit = "%"
	}

	if value > 1000000000 {
		value = value / 1000000000.0
		unit = fmt.Sprintf("G%s", unit)
	} else if value > 1000000 {
		value = value / 1000000.0
		unit = fmt.Sprintf("M%s", unit)
	} else if value > 1000 {
		value = value / 1000.0
		unit = fmt.Sprintf("K%s", unit)
	}

	formattedValue := fmt.Sprintf("%.2f", value)
	if value <= 100 {
		formattedValue = fmt.Sprintf("%.0f", value)
	}

	return formattedValue, unit
}

func dangerLevel(ps *perfstat.Perfstat) int {
	os := 0.0
	os = ps.Score("bottleneck", "cpu.*")
	os = os + ps.Score("bottleneck", "mem.*")
	os = os + ps.Score("bottleneck", "disk.*")
	os = os + ps.Score("bottleneck", "net.*")
	os = os + ps.Score("risk", "mem.*")
	os = os + ps.Score("risk", "disk.*")
	os = os + ps.Score("risk", "net.*")
	return int(math.Round((os / 7.0) * 100))
}

func addSparkline(value int, ts *signalutils.Timeseries, label string, colorize bool) (*sparkline.SparkLine, error) {
	if value != -1 {
		ts.Add(float64(value))
	}
	dangerColor := cell.ColorYellow
	if colorize {
		if value >= 80 {
			dangerColor = cell.ColorRed
		}
		if value < 20 {
			dangerColor = cell.ColorGreen
		}
	}
	sparklineDanger2, err := sparkline.New(
		sparkline.Color(dangerColor),
		sparkline.Label(label),
	)
	for _, dv := range ts.Values {
		sparklineDanger2.Add([]int{int(dv.Value)})
	}
	return sparklineDanger2, err
}
