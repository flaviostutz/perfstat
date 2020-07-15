package detectors

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		cpuPerc, err := cpu.Percent(1*time.Second, false)
		if err != nil {
			return []Issue{
				{Typ: "lib-error", Name: "cpu-overall", Message: fmt.Sprintf("Error getting cpu stats. err=%s", err)},
			}
		}

		issues := make([]Issue, 0)

		score := score(cpuPerc[0], opt.LowCPUPercRange)
		if score > 0 {

			//get hungry processes
			// processes, err := process.Processes()
			// if err != nil {
			// 	return []Issue{
			// 		{Typ: "lib-error", Name: "cpu-overall", Message: fmt.Sprintf("Error getting processes stats. err=%s", err)},
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
				Typ:  "bottleneck",
				Name: "cpu-low",
				Res: Resource{
					Typ:           "cpu",
					Name:          "cpu:overall",
					PropertyName:  "load",
					PropertyValue: cpuPerc[0],
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
