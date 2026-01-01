package collector

import (
	"net"
	"runtime"

	"github.com/shirou/gopsutil/v3/host"
	"go.uber.org/zap"
	pb "pulse/backend/pkg/external/grpc/pb"
)

const AgentVersion = "0.1.0" // TODO: Inject from build flags

func GetHostInfo() *pb.HostInfo {
	info := &pb.HostInfo{
		AgentVersion: AgentVersion,
		Arch:         runtime.GOARCH,
		Os:           runtime.GOOS,
	}

	h, err := host.Info()
	if err == nil {
		info.Hostname = h.Hostname
		info.Platform = h.Platform
		info.PlatformFamily = h.PlatformFamily
		info.PlatformVersion = h.PlatformVersion
		info.KernelVersion = h.KernelVersion
		info.UptimeSeconds = h.Uptime
	} else {
		zap.L().Warn("failed to collect host info", zap.Error(err))
	}

	// Get IPs
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					info.IpAddresses = append(info.IpAddresses, ipnet.IP.String())
				}
			}
		}
	} else {
		zap.L().Warn("failed to collect network interfaces", zap.Error(err))
	}

	return info
}
