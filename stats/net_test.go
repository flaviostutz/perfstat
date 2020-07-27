package stats

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNetStatsBasic(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ps := NewNetStats(ctx, 120*time.Second, 2*time.Second, 1)
	time.Sleep(4 * time.Second)
	assert.GreaterOrEqual(t, len(ps.NICs), 1)

	for _, n := range ps.NICs {
		_, ok := n.BytesRecv.Timeseries.Last()
		assert.True(t, ok)

		_, ok = n.PacketsRecv.Timeseries.Last()
		assert.True(t, ok)
	}
}

func TestNetTop(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ps := NewNetStats(ctx, 120*time.Second, 2*time.Second, 1)
	time.Sleep(5 * time.Second)

	td := ps.TopByteRate(true)
	assert.Greater(t, len(td), 0)
	v, ok := td[0].BytesRecv.Rate(2 * time.Second)
	assert.True(t, ok)
	assert.Less(t, v, 1000000.0)

	td = ps.TopByteRate(false)
	assert.Greater(t, len(td), 0)

	td = ps.TopPacketRate(true)
	assert.Greater(t, len(td), 0)
	v, ok = td[0].PacketsRecv.Rate(2 * time.Second)
	assert.True(t, ok)
	assert.Less(t, v, 100.0)

	td = ps.TopPacketRate(false)
	assert.Greater(t, len(td), 0)

	td = ps.TopErrorsRate(true)
	assert.Greater(t, len(td), 0)
	v, ok = td[0].ErrIn.Rate(2 * time.Second)
	assert.True(t, ok)
	assert.Less(t, v, 10.0)

	td = ps.TopErrorsRate(false)
	assert.Greater(t, len(td), 0)
}
