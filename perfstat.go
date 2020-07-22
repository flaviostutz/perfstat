package perfstat

import (
	"fmt"

	"github.com/flaviostutz/perfstat/detectors"
	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

//Perfstat performance analyser
type Perfstat struct {
	opt          detectors.Options
	cpuStats     *stats.CPUStats
	processStats *stats.ProcessStats
}

//StartPerfstat initializes a new Perfstat utility
func StartPerfstat(opt detectors.Options) Perfstat {
	p := Perfstat{
		opt: opt,
	}
	detectors.StartDetections(opt)
	return p
}

//DetectNow perform issues detection on the system once
func (p *Perfstat) DetectNow() ([]detectors.Issue, error) {
	results := make([]detectors.Issue, 0)
	if !detectors.Started {
		return []detectors.Issue{}, fmt.Errorf("Perfstat not started yet")
	}
	logrus.Debugf("Perfstat DetectNow()")
	for _, df := range detectors.DetectorFuncs {
		r := df(&p.opt)
		for _, iss := range r {
			logrus.Debugf("ISSUE: %s", iss.String())
		}
		results = append(results, r...)
	}
	return results, nil
}

func (p *Perfstat) setLogLevel(level logrus.Level) {
	logrus.SetLevel(level)
	detectors.SetLogLevel(level)
}

func (p *Perfstat) Stop() {
	detectors.StopDetections()
}
