package detectors

import (
	"fmt"
	"time"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {

		r := DetectionResult{
			Typ:  "bottleneck",
			ID:   "cpu-high-iowait",
			When: time.Now(),
		}

		iowait, ok := stats.TimeLoadPerc(&ActiveStats.CPUStats.Total.IOWait, opt.CPULoadAvgDuration)
		if !ok {
			r.Message = notEnoughDataMessage(opt.CPULoadAvgDuration)
			r.Score = -1
			return []DetectionResult{r}
		}

		r.Res = Resource{
			Typ:           "cpu",
			Name:          "cpu:all",
			PropertyName:  "iowait-perc",
			PropertyValue: iowait,
		}

		r.Score = criticityScore(iowait, opt.HighCPUWaitPercRange)

		if r.Score == 0 {
			return []DetectionResult{r}
		}

		//get most waited processes
		r.Related = make([]Resource, 0)
		for _, proc := range ActiveStats.ProcessStats.TopCPUIOWait() {
			if len(r.Related) >= 3 {
				break
			}
			iw, ok := stats.TimeLoadPerc(&proc.CPUTimes.IOWait, opt.CPULoadAvgDuration)
			if !ok {
				logrus.Tracef("Couldn't get iowait time for pid %d", proc.Pid)
				continue
			}
			res := Resource{
				Typ:           "process",
				Name:          fmt.Sprintf("%s[%d]", proc.Name, proc.Pid),
				PropertyName:  "cpu-iowait-perc",
				PropertyValue: iw,
			}
			r.Related = append(r.Related, res)
		}

		//get most waited disks
		//top read util
		for _, ds := range ActiveStats.DiskStats.TopIOUtil(true) {
			if len(r.Related) >= 2 {
				break
			}
			iw, ok := stats.TimeLoadPerc(&ds.ReadTime, opt.IORateLoadDuration)
			if !ok {
				logrus.Tracef("Couldn't get read load for disk %s", ds.Name)
				continue
			}
			if iw > 0.2 {
				res := Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", ds.Name),
					PropertyName:  "disk-read-perc",
					PropertyValue: iw,
				}
				r.Related = append(r.Related, res)
			}
		}

		//top write util
		for _, ds := range ActiveStats.DiskStats.TopIOUtil(false) {
			if len(r.Related) >= 2 {
				break
			}
			iw, ok := stats.TimeLoadPerc(&ds.ReadTime, opt.IORateLoadDuration)
			if !ok {
				logrus.Tracef("Couldn't get write load for disk %s", ds.Name)
				continue
			}
			if iw > 0.2 {
				res := Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", ds.Name),
					PropertyName:  "disk-write-perc",
					PropertyValue: iw,
				}
				r.Related = append(r.Related, res)
			}
		}

		//top read throughput
		for _, ds := range ActiveStats.DiskStats.TopByteRate(true) {
			if len(r.Related) >= 2 {
				break
			}
			iw, ok := ds.ReadBytes.Rate(opt.IORateLoadDuration)
			if !ok {
				logrus.Tracef("Couldn't get read throughput for disk %s", ds.Name)
				continue
			}
			if iw > 10000 {
				res := Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", ds.Name),
					PropertyName:  "disk-read-bps",
					PropertyValue: iw,
				}
				r.Related = append(r.Related, res)
			}
		}

		//top write throughput
		for _, ds := range ActiveStats.DiskStats.TopByteRate(false) {
			if len(r.Related) >= 2 {
				break
			}
			iw, ok := ds.WriteBytes.Rate(opt.IORateLoadDuration)
			if !ok {
				logrus.Tracef("Couldn't get write throughput for disk %s", ds.Name)
				continue
			}
			if iw > 10000 {
				res := Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", ds.Name),
					PropertyName:  "disk-write-bps",
					PropertyValue: iw,
				}
				r.Related = append(r.Related, res)
			}
		}

		//top read ops
		for _, ds := range ActiveStats.DiskStats.TopOpRate(true) {
			if len(r.Related) >= 2 {
				break
			}
			iw, ok := ds.ReadCount.Rate(opt.IORateLoadDuration)
			if !ok {
				logrus.Tracef("Couldn't get read ops for disk %s", ds.Name)
				continue
			}
			if iw > 2 {
				res := Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", ds.Name),
					PropertyName:  "disk-read-ops",
					PropertyValue: iw,
				}
				r.Related = append(r.Related, res)
			}
		}

		//top write ops
		for _, ds := range ActiveStats.DiskStats.TopOpRate(false) {
			if len(r.Related) >= 2 {
				break
			}
			iw, ok := ds.ReadCount.Rate(opt.IORateLoadDuration)
			if !ok {
				logrus.Tracef("Couldn't get write ops for disk %s", ds.Name)
				continue
			}
			if iw > 2 {
				res := Resource{
					Typ:           "disk",
					Name:          fmt.Sprintf("disk:%s", ds.Name),
					PropertyName:  "disk-write-ops",
					PropertyValue: iw,
				}
				r.Related = append(r.Related, res)
			}
		}

		return []DetectionResult{r}
	})
}
