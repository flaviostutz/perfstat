package detectors

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		cpuPercs, err := cpu.Percent(5*time.Second, true)
		if err != nil {
			return []Issue{
				{Typ: "lib-error", Message: fmt.Sprintf("Error getting cpu stats. err=%s", err)},
			}
		}

		for ci, p := range cpuPercs {
			fmt.Printf(">>> %d %f", ci, p)
			if p > opt.LowCPUCriticalPercent {

			} else if p > opt.LowCPUWarningPercent {
			}
		}

		return []Issue{
			{Typ: "TEST", Message: "TEST CPU"},
		}
	})
}
