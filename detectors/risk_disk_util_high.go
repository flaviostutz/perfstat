package detectors

import (
	"fmt"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		issues := make([]Issue, 0)

		for dname, disk := range ActiveStats.DiskStats.Disks {

			utilPerc, ok := stats.TimeLoadPerc(&disk.IoTime, opt.CPULoadAvgDuration)
			if !ok {
				logrus.Tracef("Not enough data for disk util analysis")
				return []Issue{}
			}

			score := criticityScore(utilPerc, opt.HighCPUWaitPercRange)
			logrus.Tracef("disk-util-load utilPerc=%.2f criticityScore=%.2f", utilPerc, score)

			//get processes waiting for IOs
			related := make([]Resource, 0)
			for _, proc := range ActiveStats.ProcessStats.TopCPUIOWait() {
				if len(related) >= 3 {
					break
				}
				iw, ok := stats.TimeLoadPerc(&proc.CPUTimes.IOWait, opt.CPULoadAvgDuration)
				if !ok {
					logrus.Tracef("Couldn't get iowait time for pid %d", proc.Pid)
					continue
				}
				r := Resource{
					Typ:           "process",
					Name:          fmt.Sprintf("pid:%d", proc.Pid),
					PropertyName:  "cpu-iowait-perc",
					PropertyValue: iw,
				}
				related = append(related, r)
			}

			if score > 0 {
				issues = append(issues, Issue{
					Typ:   "risk",
					ID:    "disk-high-util",
					Score: score,
					Res: Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", dname),
						PropertyName:  "util-perc",
						PropertyValue: utilPerc,
					},
					Related: related,
				})
			}

		}
		return issues
	})
}
