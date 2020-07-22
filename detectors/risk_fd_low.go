package detectors

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		to := time.Now()
		from := to.Add(-opt.IORateLoadDuration)

		usedFD, ok := ActiveStats.DiskStats.FD.UsedFD.Avg(from, to)
		if !ok {
			logrus.Tracef("Not enough data for FD analysis")
			return []Issue{}
		}
		maxFD := ActiveStats.DiskStats.FD.MaxFD
		fdUsedPerc := float64(usedFD) / float64(maxFD)

		issues := make([]Issue, 0)

		score := criticityScore(fdUsedPerc, opt.FDUsedRange)
		logrus.Tracef("fd-low load=%.2f criticityScore=%.2f", usedFD, score)
		if score > 0 {

			//get hungry processes
			related := make([]Resource, 0)
			for _, proc := range ActiveStats.ProcessStats.TopFD() {
				if len(related) >= 3 {
					break
				}
				pused, ok := proc.FD.Avg(from, to)
				if !ok {
					logrus.Tracef("Couldn't get used fd for pid %d", proc.Pid)
					continue
				}
				r := Resource{
					Typ:           "process",
					Name:          fmt.Sprintf("pid:%d", proc.Pid),
					PropertyName:  "fd-used-count",
					PropertyValue: pused,
				}
				related = append(related, r)
			}

			issues = append(issues, Issue{
				Typ:   "risk",
				ID:    "fd-low",
				Score: score,
				Res: Resource{
					Typ:           "fd",
					Name:          "fd",
					PropertyName:  "used-perc",
					PropertyValue: fdUsedPerc,
				},
				Related: related,
			})
		}

		return issues
	})
}
