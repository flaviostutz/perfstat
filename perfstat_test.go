package perfstat

import (
	"testing"
	"time"

	"github.com/flaviostutz/perfstat/detectors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var p Perfstat

func TestCPULowDetector(t *testing.T) {
	opt := detectors.NewOptions()
	//force alarms
	opt.LowCPUPercRange = [2]float64{0.0, 0.9}
	p = NewPerfstat(opt)
	p.setLogLevel(logrus.DebugLevel)
	p.Start()
	time.Sleep(3 * time.Second)

	issues, err := p.DetectNow()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(issues), 0)
	checkOneEqual(t, issues, "bottleneck", "cpu-low")
}

func checkOneEqual(t *testing.T, issues []detectors.Issue, typ string, id string) {
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
