package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollect(t *testing.T) {
	stats := Collect()
	assert.NotNil(t, stats)
	// Memory usage is almost guaranteed to be > 0 on any running system
	assert.Greater(t, stats.MemoryUsagePercent, 0.0, "memory usage should be > 0")
	// CPU usage might be 0 on idle, so we don't assert > 0 strictly, but >= 0
	assert.GreaterOrEqual(t, stats.CpuUsagePercent, 0.0)
	assert.GreaterOrEqual(t, stats.DiskUsagePercent, 0.0)
}
