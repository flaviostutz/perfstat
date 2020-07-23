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
	//force alarms
	opt.HighCPUPercRange = [2]float64{0.0, 0.9}
	opt.CPULoadAvgDuration = 1 * time.Second
	opt.IORateLoadDuration = 1 * time.Second
	p := Start(opt)
	p.SetLogLevel(logrus.DebugLevel)
	time.Sleep(6 * time.Second)

	for _, is := range p.curResults {
		fmt.Printf("]]]] %s\n", is.String())
	}

	crit := p.TopCriticity()
	for _, is := range crit {
		fmt.Printf(">>>> %s\n", is.String())
	}
	fmt.Printf("\n\n\n")
	time.Sleep(1 * time.Second)
	for _, is := range crit {
		fmt.Printf(">>>> %s\n", is.String())
	}
	fmt.Printf("\n\n\n")
	time.Sleep(1 * time.Second)
	for _, is := range crit {
		fmt.Printf(">>>> %s\n", is.String())
	}
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
