package stats

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemBasic(t *testing.T) {
	// logrus.SetLevel(logrus.DebugLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewMemStats(ctx, 60*time.Second, 2)
	time.Sleep(5 * time.Second)

	tot := s.Total
	assert.GreaterOrEqualf(t, tot, uint64(1000), "")

	v, ok := s.Free.Last()
	assert.True(t, ok)
	assert.GreaterOrEqualf(t, v.Value, 1000.0, "")

	v, ok = s.Used.Last()
	assert.True(t, ok)
	assert.GreaterOrEqualf(t, v.Value, 1000.0, "")

	v, ok = s.Available.Last()
	assert.True(t, ok)
	assert.GreaterOrEqualf(t, v.Value, 1000.0, "")
}
