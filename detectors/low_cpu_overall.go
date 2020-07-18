package detectors

import (
	"fmt"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		idle, ok := stats.CPUAvgPerc(&ActiveStats.CPUStats.Total.Idle, opt.CPULoadAvgDuration)
		if !ok {
			logrus.Debugf("Not enough data for CPU analysis")
			return []Issue{}
		}

		load := 1.0 - idle

		issues := make([]Issue, 0)

		score := criticityScore(load, opt.LowCPUPercRange)
		// logrus.Debugf("criticityScore=%.2f overall-cpu-load=%.2f", score, load)
		if score > 0 {

			//get hungry processes
			related := make([]Resource, 0)
			for _, proc := range ActiveStats.ProcessStats.GetTopCPULoad() {
				if len(related) >= 3 {
					break
				}
				ps, _ := stats.CPUAvgPerc(&proc.CPUTimes.System, opt.CPULoadAvgDuration)
				pu, _ := stats.CPUAvgPerc(&proc.CPUTimes.User, opt.CPULoadAvgDuration)
				r := Resource{
					Typ:           "process",
					Name:          fmt.Sprintf("pid:%d", proc.Pid),
					PropertyName:  "load",
					PropertyValue: pu + ps,
				}
				related = append(related, r)
			}

			issues = append(issues, Issue{
				Typ:   "bottleneck",
				ID:    "cpu-low",
				Score: score,
				Res: Resource{
					Typ:           "cpu",
					Name:          "cpu:overall",
					PropertyName:  "load",
					PropertyValue: load,
				},
				Related: related,
			})
		}

		return issues
	})
}
