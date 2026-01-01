package collector

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
	pb "pulse/backend/pkg/external/grpc/pb"
)

// Collect gathers system metrics (CPU, RAM, Disk).
// It blocks for a short duration (100ms) to calculate CPU usage.
func Collect() *pb.SystemStats {
	stats := &pb.SystemStats{}

	// CPU Usage
	// We wait 100ms to get a sample.
	c, err := cpu.Percent(100*time.Millisecond, false)
	if err == nil && len(c) > 0 {
		stats.CpuUsagePercent = c[0]
	} else {
		zap.L().Warn("failed to collect cpu stats", zap.Error(err))
	}

	// Memory Usage
	v, err := mem.VirtualMemory()
	if err == nil {
		stats.MemoryUsagePercent = v.UsedPercent
	} else {
		zap.L().Warn("failed to collect memory stats", zap.Error(err))
	}

	// Disk Usage
	// We check the partition hosting the current working directory.
	d, err := disk.Usage(".")
	if err == nil {
		stats.DiskUsagePercent = d.UsedPercent
	} else {
		// Fallback to root if "." fails
		d, err = disk.Usage("/")
		if err == nil {
			stats.DiskUsagePercent = d.UsedPercent
		} else {
			zap.L().Warn("failed to collect disk stats", zap.Error(err))
		}
	}

	return stats
}
