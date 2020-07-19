package detectors

func init() {
	RegisterDetector(func(opt *Options) []Issue {
		// parei aqui
		// used, ok := ActiveStats.ProcessStats.TopFD()MemStats.Used.Last()
		// if !ok {
		// 	logrus.Debugf("Not enough data for Memory analysis")
		// 	return []Issue{}
		// }

		issues := make([]Issue, 0)

		// total, ok := ActiveStats.MemStats.Total.Last()

		// load := used.Value / total.Value
		// score := criticityScore(load, opt.HighMemPercRange)
		// logrus.Tracef("mem-low load=%.2f criticityScore=%.2f", load, score)
		// if score > 0 {

		// 	//get hungry processes
		// 	related := make([]Resource, 0)
		// 	for _, proc := range ActiveStats.ProcessStats.TopMemUsed() {
		// 		if len(related) >= 5 {
		// 			break
		// 		}
		// 		mt, ok := proc.MemoryTotal.Last()
		// 		if !ok {
		// 			logrus.Tracef("Couldn't get memory total for pid %d", proc.Pid)
		// 			continue
		// 		}
		// 		r := Resource{
		// 			Typ:           "process",
		// 			Name:          fmt.Sprintf("pid:%d", proc.Pid),
		// 			PropertyName:  "mem-used-bytes",
		// 			PropertyValue: mt.Value,
		// 		}
		// 		related = append(related, r)
		// 	}

		// 	issues = append(issues, Issue{
		// 		Typ:   "bottleneck",
		// 		ID:    "mem-low",
		// 		Score: score,
		// 		Res: Resource{
		// 			Typ:           "mem",
		// 			Name:          "ram",
		// 			PropertyName:  "used-perc",
		// 			PropertyValue: load,
		// 		},
		// 		Related: related,
		// 	})
		// }

		return issues
	})
}
