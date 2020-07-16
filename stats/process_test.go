package stats

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestProcessStatsBasic(t *testing.T) {
	// logrus.SetLevel(logrus.DebugLevel)
	ps := NewProcessStats(60*time.Second, 0.1)
	time.Sleep(40 * time.Second)
	for _, p := range ps.Processes {
		logrus.Infof(">>>> %d", p.CPUTimes.User.Size())
		tc, ok := CPUAvgPerc(&p.CPUTimes.User, 3*time.Second)
		assert.True(t, ok)
		assert.LessOrEqualf(t, tc, 1.0, "")
		return
	}
}
