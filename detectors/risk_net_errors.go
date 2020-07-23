package detectors

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []DetectionResult {

		issues := make([]DetectionResult, 0)

		for nname, nic := range ActiveStats.NetStats.NICs {

			//NIC ERRORS IN
			r := DetectionResult{
				Typ:  "risk",
				ID:   "net-high-errin",
				When: time.Now(),
			}

			errRate, ok := nic.ErrIn.Rate(opt.IORateLoadDuration)
			if !ok {
				r.Message = notEnoughDataMessage(opt.IORateLoadDuration)
				return []DetectionResult{r}
			}

			r.Res = Resource{
				Typ:           "net",
				Name:          fmt.Sprintf("nic:%s", nname),
				PropertyName:  "errin-pps",
				PropertyValue: errRate,
			}

			r.Score = criticityScore(errRate, opt.NICErrorsRange)

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetErrRate(true) {
					if len(r.Related) >= 3 {
						break
					}
					perr, ok := proc.TotalNetIOCounters.ErrIn.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Tracef("Couldn't get net err in for pid %d", proc.Pid)
						continue
					}
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-errin-pps",
						PropertyValue: perr,
					}
					r.Related = append(r.Related, res)
				}
			}
			issues = append(issues, r)

			//NIC ERRORS OUT
			r = DetectionResult{
				Typ:  "risk",
				ID:   "net-high-errout",
				When: time.Now(),
			}

			errRate, ok = nic.ErrOut.Rate(opt.IORateLoadDuration)
			if !ok {
				r.Message = notEnoughDataMessage(opt.IORateLoadDuration)
				return []DetectionResult{r}
			}

			r.Res = Resource{
				Typ:           "net",
				Name:          fmt.Sprintf("nic:%s", nname),
				PropertyName:  "err-out-pps",
				PropertyValue: errRate,
			}

			r.Score = criticityScore(errRate, opt.NICErrorsRange)

			if r.Score > 0 {
				//get hungry processes
				r.Related = make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetErrRate(false) {
					if len(r.Related) >= 3 {
						break
					}
					perr, ok := proc.TotalNetIOCounters.ErrOut.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Tracef("Couldn't get net err out for pid %d", proc.Pid)
						continue
					}
					res := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-errout-pps",
						PropertyValue: perr,
					}
					r.Related = append(r.Related, res)
				}

			}
			issues = append(issues, r)
		}
		return issues
	})
}
