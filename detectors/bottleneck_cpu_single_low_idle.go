package detectors

import (
	"fmt"
	"time"

	"github.com/flaviostutz/perfstat/stats"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {

		issues := make([]DetectionResult, 0)

		for cpui, cpu := range ActiveStats.CPUStats.CPU {

			r := DetectionResult{
				Typ:  "bottleneck",
				ID:   "cpu-single-low-idle",
				When: time.Now(),
			}

			idle, ok := stats.TimeLoadPerc(&cpu.Idle, opt.CPULoadAvgDuration)
			if !ok {
				r.Message = notEnoughDataMessage(opt.CPULoadAvgDuration)
				return []DetectionResult{r}
			}

			load := 1.0 - idle
			r.Score = criticityScore(load, opt.HighCPUPercRange)

			r.Res = Resource{
				Typ:           "cpu",
				Name:          fmt.Sprintf("cpu:%d", cpui),
				PropertyName:  "load-perc",
				PropertyValue: load,
			}

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopCPULoad() {
					if len(r.Related) >= 3 {
						break
					}
					ps, _ := stats.TimeLoadPerc(&proc.CPUTimes.System, opt.CPULoadAvgDuration)
					pu, _ := stats.TimeLoadPerc(&proc.CPUTimes.User, opt.CPULoadAvgDuration)
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "cpu-all-load-perc",
						PropertyValue: pu + ps,
					}
					r.Related = append(r.Related, res)
				}
			}

			issues = append(issues, r)
		}
		return issues
	})
}
