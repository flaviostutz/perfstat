package detectors

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		used, ok := ActiveStats.MemStats.Used.Last()
		if !ok {
			logrus.Debugf("Not enough data for Memory analysis")
			return []Issue{}
		}

		issues := make([]Issue, 0)

		total := ActiveStats.MemStats.Total

		load := used.Value / float64(total)
		score := criticityScore(load, opt.HighMemPercRange)
		logrus.Tracef("mem-low load=%.2f criticityScore=%.2f", load, score)
		if score > 0 {

			//get hungry processes
			related := make([]Resource, 0)
			for _, proc := range ActiveStats.ProcessStats.TopMemUsed() {
				if len(related) >= 5 {
					break
				}
				mt, ok := proc.MemoryTotal.Last()
				if !ok {
					logrus.Tracef("Couldn't get memory total for pid %d", proc.Pid)
					continue
				}
				r := Resource{
					Typ:           "process",
					Name:          fmt.Sprintf("pid:%d", proc.Pid),
					PropertyName:  "mem-used-bytes",
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
					PropertyName:  "used-perc",
					PropertyValue: load,
				},
				Related: related,
			})
		}

		return issues
	})
}
