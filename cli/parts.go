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
	c1, err := button.New(fmt.Sprintf("[%d] %s (%s)", bvalue, blabel, bKeyText),
		func() error { return nil },
		button.GlobalKey(bKey),
		button.Width(bwidth),
		button.Height(bheight),
		button.FillColor(cell.ColorYellow),
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

	idx2 := strings.LastIndex(dr.Res.PropertyName, "-")
	if idx2 == -1 {
		return "ERROR"
	}
	unit := dr.Res.PropertyName[idx2+1:]

	value := dr.Res.PropertyValue

	if unit == "bps" {
		unit = "Bps"
	} else if unit == "perc" {
		value = dr.Res.PropertyValue * 100
		unit = "%"
	}

	if value > 1000 {
		value = value / 1000
		unit = fmt.Sprintf("K%s", unit)
	} else if value > 1000000 {
		value = value / 1000000
		unit = fmt.Sprintf("M%s", unit)
	}

	valueStr := fmt.Sprintf("%.2f", value)
	if value <= 100 {
		valueStr = fmt.Sprintf("%.0f", value)
	}

	return fmt.Sprintf("%d %s %s=%s%s", perc(dr.Score), idn, dr.Res.Name, valueStr, unit)
}

func perc(v float64) int {
	return int(math.Round(v * 100))
}
