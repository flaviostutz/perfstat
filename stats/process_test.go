package stats

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestProcessStatsBasic(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ps := NewProcessStats(120*time.Second, 1)
	time.Sleep(4 * time.Second)
	assert.GreaterOrEqual(t, len(ps.Processes), 1)
	for _, p := range ps.Processes {
		if strings.Contains(p.Name, "test") || strings.Contains(p.Name, "go") {
			continue
		}
		assert.GreaterOrEqual(t, p.CPUTimes.User.Size(), 3)
		assert.GreaterOrEqual(t, p.IOCounters.ReadBytes.Timeseries.Size(), 3)
		assert.GreaterOrEqual(t, p.MemoryTotal.Size(), 3)
		tc, ok := CPUAvgPerc(&p.CPUTimes.User, 1*time.Second)
		assert.True(t, ok, fmt.Sprintf("name=%s pid=%d", p.Name, p.Pid))
		assert.LessOrEqualf(t, tc, 1.0, fmt.Sprintf("name=%s pid=%d", p.Name, p.Pid))
	}
}
