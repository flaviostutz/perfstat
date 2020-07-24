package perfstat

import (
	"fmt"
	"testing"
	"time"

	"github.com/flaviostutz/perfstat/detectors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var p Perfstat

func TestDetectNow(t *testing.T) {
	opt := detectors.NewOptions()
	//force alarms
	opt.HighCPUPercRange = [2]float64{0.0, 0.9}
	opt.CPULoadAvgDuration = 1 * time.Second
	opt.IORateLoadDuration = 1 * time.Second
	p := Start(opt)
	p.SetLogLevel(logrus.DebugLevel)
	time.Sleep(6 * time.Second)

	issues, err := p.DetectNow()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(issues), 0)
	checkOneEqual(t, issues, "bottleneck", "cpu-low")
}

func TestRollingDetections(t *testing.T) {
	opt := detectors.NewOptions()
	opt.CPULoadAvgDuration = 10 * time.Second
	opt.IORateLoadDuration = 10 * time.Second
	opt.IOLimitsSpan = 30 * time.Second
	opt.MemAvgDuration = 30 * time.Second
	opt.MemLeakDuration = 30 * time.Second

	p := Start(opt)
	p.SetLogLevel(logrus.DebugLevel)
	time.Sleep(6 * time.Second)

	for {
		fmt.Printf("\n\n\n")
		crit := p.TopCriticity(0.01, "", "", false)
		for _, is := range crit {
			// fmt.Printf(">>>> %s\n", is.String())
			fmt.Printf("%.2f %s [%s %s=%.2f] (%s)\n", is.Score, is.ID, is.Res.Name, is.Res.PropertyName, is.Res.PropertyValue, is.Typ)
		}
		time.Sleep(5 * time.Second)
	}
	// checkOneEqual(t, crit, "bottleneck", "cpu-low-idle")
	// checkOneEqual(t, crit, "risk", "disk-low-space")
	// checkOneEqual(t, crit, "risk", "disk-high-util")
	// checkOneEqual(t, crit, "bottleneck", "cpu-high-iowait")
	// checkOneEqual(t, crit, "bottleneck", "mem-low")

	// fmt.Printf("\n\n\nCPU\n")
	// time.Sleep(1 * time.Second)
	// crit = p.TopCriticity(0, "", "cpu", true)
	// for _, is := range crit {
	// 	fmt.Printf(">>>> %s\n", is.String())
	// }
	// fmt.Printf("\n\n\nMEM\n")
	// time.Sleep(1 * time.Second)
	// crit = p.TopCriticity(0, "", "mem", true)
	// for _, is := range crit {
	// 	fmt.Printf(">>>> %s\n", is.String())
	// }
	// fmt.Printf("\n\n\nNET\n")
	// time.Sleep(1 * time.Second)
	// crit = p.TopCriticity(0, "", "net", true)
	// for _, is := range crit {
	// 	fmt.Printf(">>>> %s\n", is.String())
	// }
	// fmt.Printf("\n\n\nDISK\n")
	// time.Sleep(1 * time.Second)
	// crit = p.TopCriticity(0, "", "disk", true)
	// for _, is := range crit {
	// 	fmt.Printf(">>>> %s\n", is.String())
	// }
}

func checkOneEqual(t *testing.T, issues []detectors.DetectionResult, typ string, id string) {
	found := false
	for _, is := range issues {
		if typ != "" && id != "" {
			if is.ID == id && is.Typ == typ {
				found = true
				break
			}

		}
		if typ != "" && id == "" {
			if is.Typ == typ {
				found = true
				break
			}
		}
		if typ == "" && id != "" {
			if is.ID == id {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("No issues found with type='%s', id='%s'", typ, id)
	}
}
