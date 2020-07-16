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

func init() {
	p = NewPerfstat(detectors.NewOptions())
	p.setLogLevel(logrus.DebugLevel)
	p.Start()
	time.Sleep(3 * time.Second)
}

func TestBasic(t *testing.T) {
	issues, err := p.DetectNow()
	assert.Nil(t, err)
	for _, v := range issues {
		fmt.Printf("FOUND ISSUE: %s", v.String())
	}
}
