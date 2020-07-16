package detectors

import (
	"time"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		idle, ok := stats.CPUAvgPerc(&ActiveStats.CPUStats.Total.Idle, 1*time.Second)
		if !ok {
			logrus.Tracef("Not enough data for CPU analysis")
			return []Issue{}
		}

		load := 1.0 - idle

		issues := make([]Issue, 0)

		score := criticityScore(load, opt.LowCPUPercRange)
		// logrus.Debugf("criticityScore=%.2f overall-cpu-load=%.2f", score, load)
		if score > 0 {

			//get hungry processes
			// processes, err := process.Processes()
			// if err != nil {
			// 	return []Issue{
			// 		{Typ: "internal-error", Name: "cpu-overall", Message: fmt.Sprintf("Error getting processes stats. err=%s", err)},
			// 	}
			// }

			// for _, proc := range processes {
			// 	pperc, err := proc.Percent(1 * time.Second)
			// 	if err != nil {
			// 		return []Issue{
			// 			{Typ: "lib-error", Name: "cpu-overall", Message: fmt.Sprintf("Error getting processes stats. err=%s", err)},
			// 		}
			// 	}
			// }

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
				Related: []Resource{
					// {
					// 	Typ:           "top-process",
					// 	Name:          fmt.Sprintf("cpu:%d", ci),
					// 	PropertyName:  "load",
					// 	PropertyValue: p1,
					// },
				},
				// Message string
			})
		}

		return issues
	})
}
