package stats

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestProcessStatsBasic(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ps := NewProcessStats(ctx, 120*time.Second, 1*time.Second, 1*time.Second, 1*time.Second, 1.0)
	time.Sleep(7 * time.Second)
	assert.GreaterOrEqual(t, len(ps.Processes), 1)
	for _, p := range ps.Processes {
		if strings.Contains(p.Name, "test") || strings.Contains(p.Name, "go") {
			continue
		}
		assert.GreaterOrEqual(t, p.CPUTimes.User.Size(), 3)
		assert.GreaterOrEqual(t, p.IOCounters.ReadBytes.Timeseries.Size(), 3)
		assert.GreaterOrEqual(t, p.MemoryTotal.Size(), 3)
		tc, ok := TimeLoadPerc(&p.CPUTimes.User, 1*time.Second)
		assert.True(t, ok, fmt.Sprintf("name=%s pid=%d", p.Name, p.Pid))
		assert.LessOrEqualf(t, tc, 1.0, fmt.Sprintf("name=%s pid=%d", p.Name, p.Pid))
	}
}
