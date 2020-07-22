package detectors

import (
	"fmt"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {

		issues := make([]Issue, 0)

		for cpui, cpu := range ActiveStats.CPUStats.CPU {
			idle, ok := stats.TimeLoadPerc(&cpu.Idle, opt.CPULoadAvgDuration)
			if !ok {
				logrus.Debugf("Not enough data for single CPU analysis")
				return []Issue{}
			}

			load := 1.0 - idle
			score := criticityScore(load, opt.HighCPUPercRange)
			logrus.Tracef("cpu-single-low load=%.2f criticityScore=%.2f", load, score)

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
						PropertyName:  "cpu-all-load-perc",
						PropertyValue: pu + ps,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "bottleneck",
					ID:    "cpu-single-low-idle",
					Score: score,
					Res: Resource{
						Typ:           "cpu",
						Name:          fmt.Sprintf("cpu:%d", cpui),
						PropertyName:  "load-perc",
						PropertyValue: load,
					},
					Related: related,
				})
			}
		}
		return issues
	})
}
