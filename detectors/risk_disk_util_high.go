package detectors

import (
	"fmt"
	"time"

	"github.com/flaviostutz/perfstat/stats"
	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {

		issues := make([]DetectionResult, 0)

		for dname, disk := range ActiveStats.DiskStats.Disks {
			r := DetectionResult{
				Typ:  "risk",
				ID:   "disk-high-util",
				When: time.Now(),
			}

			utilPerc, ok := stats.TimeLoadPerc(&disk.IoTime, opt.CPULoadAvgDuration)
			if !ok {
				r.Message = notEnoughDataMessage(opt.CPULoadAvgDuration)
				r.Score = -1
				return []DetectionResult{r}
			}

			r.Res = Resource{
				Typ:           "disk",
				Name:          fmt.Sprintf("disk:%s", dname),
				PropertyName:  "util-perc",
				PropertyValue: utilPerc,
			}

			cc, err := ActiveStats.CPUStats.CPUCount()
			if err != nil {
				r.Message = fmt.Sprintf("Couldn't get CPU count. err=%s", err)
				r.Score = -1
				return []DetectionResult{r}
			}
			ranges := [2]float64{opt.HighDiskUtilPercRange[0] * float64(cc), opt.HighDiskUtilPercRange[1] * float64(cc)}
			r.Score = criticityScore(utilPerc, ranges)

			//get processes waiting for IOs
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
			issues = append(issues, r)
		}
		return issues
	})
}
