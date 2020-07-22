package detectors

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		issues := make([]Issue, 0)

		for nname, nic := range ActiveStats.NetStats.NICs {

			//NIC ERRORS IN
			errRate, ok := nic.ErrIn.Rate(opt.IORateLoadDuration)
			if !ok {
				logrus.Tracef("Not enough data for net errors in analysis")
				return []Issue{}
			}

			score := criticityScore(errRate, opt.NICErrorsRange)
			logrus.Tracef("net-errors-in errInRate=%.2f criticityScore=%.2f", errRate, score)

			if score > 0 {

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetErrRate(true) {
					if len(related) >= 3 {
						break
					}
					perr, ok := proc.TotalNetIOCounters.ErrIn.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Tracef("Couldn't get net err in for pid %d", proc.Pid)
						continue
					}
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-errin-pps",
						PropertyValue: perr,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "risk",
					ID:    "net-high-errin",
					Score: score,
					Res: Resource{
						Typ:           "net",
						Name:          fmt.Sprintf("nic:%s", nname),
						PropertyName:  "errin-pps",
						PropertyValue: errRate,
					},
					Related: related,
				})
			}

			//NIC ERRORS OUT
			errRate, ok = nic.ErrOut.Rate(opt.IORateLoadDuration)
			if !ok {
				logrus.Tracef("Not enough data for net errors out analysis")
				return []Issue{}
			}

			score = criticityScore(errRate, opt.NICErrorsRange)
			logrus.Tracef("net-errors-out errOutRate=%.2f criticityScore=%.2f", errRate, score)

			if score > 0 {

				//get hungry processes
				related := make([]Resource, 0)
				for _, proc := range ActiveStats.ProcessStats.TopNetErrRate(false) {
					if len(related) >= 3 {
						break
					}
					perr, ok := proc.TotalNetIOCounters.ErrOut.Rate(opt.IORateLoadDuration)
					if !ok {
						logrus.Tracef("Couldn't get net err out for pid %d", proc.Pid)
						continue
					}
					r := Resource{
						Typ:           "process",
						Name:          fmt.Sprintf("pid:%d", proc.Pid),
						PropertyName:  "net-errout-pps",
						PropertyValue: perr,
					}
					related = append(related, r)
				}

				issues = append(issues, Issue{
					Typ:   "risk",
					ID:    "net-high-errout",
					Score: score,
					Res: Resource{
						Typ:           "net",
						Name:          fmt.Sprintf("nic:%s", nname),
						PropertyName:  "err-out-pps",
						PropertyValue: errRate,
					},
				})
			}

		}
		return issues
	})
}
