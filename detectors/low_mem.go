package detectors

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		used, ok := ActiveStats.MemStats.Used.GetLastValue()
		if !ok {
			logrus.Debugf("Not enough data for Memory analysis")
			return []Issue{}
		}

		issues := make([]Issue, 0)

		total, ok := ActiveStats.MemStats.Total.GetLastValue()

		load := used.Value / total.Value
		score := criticityScore(load, opt.LowMemPercRange)
		logrus.Debugf("criticityScore=%.2f mem-load=%.2f", score, load)
		if score > 0 {

			//get hungry processes
			related := make([]Resource, 0)
			for _, proc := range ActiveStats.ProcessStats.GetTopMem() {
				if len(related) >= 5 {
					break
				}
				mt, ok := proc.MemoryTotal.GetLastValue()
				if !ok {
					continue
				}
				r := Resource{
					Typ:           "process",
					Name:          fmt.Sprintf("pid:%d", proc.Pid),
					PropertyName:  "mem",
					PropertyValue: mt.Value,
				}
				related = append(related, r)
			}

			issues = append(issues, Issue{
				Typ:   "bottleneck",
				ID:    "mem-low",
				Score: score,
				Res: Resource{
					Typ:           "mem",
					Name:          "ram",
					PropertyName:  "load",
					PropertyValue: load,
				},
				Related: related,
			})
		}

		return issues
	})
}
