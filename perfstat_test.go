package perfstat

import (
	"fmt"
	"testing"

	"github.com/flaviostutz/perfstat/detectors"
)

func TestBasic(t *testing.T) {
	p := NewPerfstat(detectors.NewOptions())
	issues := p.DetectNow()
	for _, v := range issues {
		fmt.Printf("%v", v)
	}
}
