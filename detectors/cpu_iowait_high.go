package detectors

import (
	"fmt"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		iowait, ok := stats.TimeLoadPerc(&ActiveStats.CPUStats.Total.IOWait, opt.CPULoadAvgDuration)
		if !ok {
			logrus.Debugf("Not enough data for CPU iowait analysis")
			return []Issue{}
		}

		issues := make([]Issue, 0)

		score := criticityScore(iowait, opt.HighCPUWaitPercRange)

		logrus.Tracef("cpu-iowait-high iowait=%.2f criticityScore=%.2f", iowait, score)
		if score > 0 {

			//get most waited processes
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

			//get most waited disks
			//top read util
			for _, ds := range ActiveStats.DiskStats.TopIOUtil(true) {
				if len(related) >= 2 {
					break
				}
				iw, ok := stats.TimeLoadPerc(&ds.ReadTime, opt.IORateLoadDuration)
				if !ok {
					logrus.Tracef("Couldn't get read load for disk %s", ds.Name)
					continue
				}
				if iw > 0.2 {
					r := Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", ds.Name),
						PropertyName:  "disk-read-perc",
						PropertyValue: iw,
					}
					related = append(related, r)
				}
			}

			//top write util
			for _, ds := range ActiveStats.DiskStats.TopIOUtil(false) {
				if len(related) >= 2 {
					break
				}
				iw, ok := stats.TimeLoadPerc(&ds.ReadTime, opt.IORateLoadDuration)
				if !ok {
					logrus.Tracef("Couldn't get write load for disk %s", ds.Name)
					continue
				}
				if iw > 0.2 {
					r := Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", ds.Name),
						PropertyName:  "disk-write-perc",
						PropertyValue: iw,
					}
					related = append(related, r)
				}
			}

			//top read throughput
			for _, ds := range ActiveStats.DiskStats.TopByteRate(true) {
				if len(related) >= 2 {
					break
				}
				iw, ok := ds.ReadBytes.Rate(opt.IORateLoadDuration)
				if !ok {
					logrus.Tracef("Couldn't get read throughput for disk %s", ds.Name)
					continue
				}
				if iw > 10000 {
					r := Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", ds.Name),
						PropertyName:  "disk-read-bps",
						PropertyValue: iw,
					}
					related = append(related, r)
				}
			}

			//top write throughput
			for _, ds := range ActiveStats.DiskStats.TopByteRate(false) {
				if len(related) >= 2 {
					break
				}
				iw, ok := ds.WriteBytes.Rate(opt.IORateLoadDuration)
				if !ok {
					logrus.Tracef("Couldn't get write throughput for disk %s", ds.Name)
					continue
				}
				if iw > 10000 {
					r := Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", ds.Name),
						PropertyName:  "disk-write-bps",
						PropertyValue: iw,
					}
					related = append(related, r)
				}
			}

			//top read ops
			for _, ds := range ActiveStats.DiskStats.TopOpRate(true) {
				if len(related) >= 2 {
					break
				}
				iw, ok := ds.ReadCount.Rate(opt.IORateLoadDuration)
				if !ok {
					logrus.Tracef("Couldn't get read ops for disk %s", ds.Name)
					continue
				}
				if iw > 2 {
					r := Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", ds.Name),
						PropertyName:  "disk-read-ops",
						PropertyValue: iw,
					}
					related = append(related, r)
				}
			}

			//top write ops
			for _, ds := range ActiveStats.DiskStats.TopOpRate(false) {
				if len(related) >= 2 {
					break
				}
				iw, ok := ds.ReadCount.Rate(opt.IORateLoadDuration)
				if !ok {
					logrus.Tracef("Couldn't get write ops for disk %s", ds.Name)
					continue
				}
				if iw > 2 {
					r := Resource{
						Typ:           "disk",
						Name:          fmt.Sprintf("disk:%s", ds.Name),
						PropertyName:  "disk-write-ops",
						PropertyValue: iw,
					}
					related = append(related, r)
				}
			}

			issues = append(issues, Issue{
				Typ:   "bottleneck",
				ID:    "cpu-high-iowait",
				Score: score,
				Res: Resource{
					Typ:           "cpu",
					Name:          "cpu:all",
					PropertyName:  "iowait-perc",
					PropertyValue: iowait,
				},
				Related: related,
			})
		}

		return issues
	})
}
