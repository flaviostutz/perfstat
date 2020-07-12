package perfstat

import (
	"github.com/flaviostutz/perfstat/detectors"
	_ "github.com/flaviostutz/perfstat/detectors"
)

//Perfstat performance analyser
type Perfstat struct {
	opt detectors.Options
}

//NewPerfstat initializes a new Perfstat utility
func NewPerfstat(opt detectors.Options) Perfstat {
	return Perfstat{
		opt: opt,
	}
}

//DetectNow perform issues detection on the system once
func (p *Perfstat) DetectNow() []detectors.Issue {
	results := make([]detectors.Issue, 0)
	for _, df := range detectors.DetectorFuncs {
		r := df(&p.opt)
		results = append(results, r...)
	}
	return results
}
