package detectors

import (
	"fmt"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		idle, ok := stats.TimeLoadPerc(&ActiveStats.CPUStats.Total.Idle, opt.CPULoadAvgDuration)
		if !ok {
			logrus.Debugf("Not enough data for CPU analysis")
			return []Issue{}
		}

		load := 1.0 - idle

		issues := make([]Issue, 0)

		score := criticityScore(load, opt.HighCPUPercRange)
		logrus.Tracef("cpu-low load=%.2f criticityScore=%.2f", load, score)
		if score > 0 {

			//get hungry processes
			related := make([]Resource, 0)
			for _, proc := range ActiveStats.ProcessStats.TopCPULoad() {
				if len(related) >= 3 {
					break
				}
				ps, _ := stats.TimeLoadPerc(&proc.CPUTimes.System, opt.CPULoadAvgDuration)
				pu, _ := stats.TimeLoadPerc(&proc.CPUTimes.User, opt.CPULoadAvgDuration)
				r := Resource{
					Typ:           "process",
					Name:          fmt.Sprintf("pid:%d", proc.Pid),
					PropertyName:  "cpu-load-perc",
					PropertyValue: pu + ps,
				}
				related = append(related, r)
			}

			//check if vm hypervisor is stealing CPU power
			steal, ok := stats.TimeLoadPerc(&ActiveStats.CPUStats.Total.Steal, opt.CPULoadAvgDuration)
			if ok && steal > 0.2 {
				r := Resource{
					Typ:           "cpu",
					Name:          "cpu:total",
					PropertyName:  "cpu-steal-perc",
					PropertyValue: steal,
				}
				related = append(related, r)
			}

			issues = append(issues, Issue{
				Typ:   "bottleneck",
				ID:    "cpu-low-idle",
				Score: score,
				Res: Resource{
					Typ:           "cpu",
					Name:          "cpu:all",
					PropertyName:  "load-perc",
					PropertyValue: load,
				},
				Related: related,
			})
		}

		return issues
	})
}
