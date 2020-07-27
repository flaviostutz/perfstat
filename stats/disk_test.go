package stats

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDiskStatsBasic(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ps := NewDiskStats(ctx, 120*time.Second, 2*time.Second, 1)
	time.Sleep(4 * time.Second)
	assert.GreaterOrEqual(t, len(ps.Disks), 1)

	for _, p := range ps.Disks {
		rl, ok := p.ReadTime.Last()
		assert.True(t, ok)
		assert.Greater(t, rl.Value, 0.0)

		rl, ok = p.ReadBytes.Timeseries.Last()
		assert.True(t, ok)
		assert.Greater(t, rl.Value, 0.0)

		rl, ok = p.ReadCount.Timeseries.Last()
		assert.True(t, ok)
		assert.Greater(t, rl.Value, 0.0)
	}
}

func TestDiskTop(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ps := NewDiskStats(ctx, 120*time.Second, 2*time.Second, 1)
	time.Sleep(5 * time.Second)

	td := ps.TopByteRate(true)
	assert.Greater(t, len(td), 0)
	v, ok := td[0].ReadBytes.Rate(2 * time.Second)
	assert.True(t, ok)
	assert.Less(t, v, 1000000.0)

	td = ps.TopByteRate(false)
	assert.Greater(t, len(td), 0)

	td = ps.TopOpRate(true)
	assert.Greater(t, len(td), 0)

	td = ps.TopOpRate(false)
	assert.Greater(t, len(td), 0)

	td = ps.TopIOUtil(true)
	assert.Greater(t, len(td), 0)
	v, ok = TimeLoadPerc(&td[0].ReadTime, 2*time.Second)
	assert.True(t, ok)
	assert.Less(t, v, 1.0)

	td = ps.TopIOUtil(false)
	assert.Greater(t, len(td), 0)
}
