package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "pulse/backend/pkg/external/grpc/pb"
	"pulse/backend/pkg/external/logging"
)

var (
	serverAddr = flag.String("server", "localhost:50051", "Pulse server address")
	caCert     = flag.String("ca", "certs/ca.crt", "Path to CA certificate")
	certFile   = flag.String("cert", "certs/agent.crt", "Path to agent certificate")
	keyFile    = flag.String("key", "certs/agent.key", "Path to agent key")
	token      = flag.String("token", "", "Install token (for initial registration)")
)

type Agent struct {
	client     pb.AgentServiceClient
	conn       *grpc.ClientConn
	identity   string
	logger     *zap.Logger
	stream     pb.AgentService_ConnectClient
	shutdownCh chan struct{}
}

func main() {
	flag.Parse()

	// Initialize logger (simplified for agent)
	logger, _ := zap.NewProduction()
	logging.Logger = logger

	agent := &Agent{
		logger:     logger,
		shutdownCh: make(chan struct{}),
	}

	// Connect to server
	if err := agent.connect(); err != nil {
		logger.Fatal("Failed to connect to server", zap.Error(err))
	}
	defer agent.conn.Close()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start main loops
	go agent.heartbeatLoop()
	go agent.metricsLoop()
	go agent.taskLoop()

	<-sigCh
	logger.Info("Shutting down agent...")
	close(agent.shutdownCh)
	time.Sleep(1 * time.Second) // Give loops time to finish
}

func (a *Agent) connect() error {
	// Load TLS credentials
	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		// If certs don't exist and we have a token, we might need to register first
		// But for this MVP, we assume certs are pre-provisioned or generated via separate process
		// To keep it simple, we'll just fail if certs are missing,
		// implying a "pulse-agent register" command would be run separately or certs placed manually.
		return fmt.Errorf("failed to load client certs: %v", err)
	}

	caBody, err := os.ReadFile(*caCert)
	if err != nil {
		return fmt.Errorf("failed to read CA cert: %v", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caBody) {
		return fmt.Errorf("failed to append CA cert")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ServerName:   "pulse-server", // Must match server cert CN
	})

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}

	a.conn = conn
	a.client = pb.NewAgentServiceClient(conn)

	// Determine identity from cert (or just hostname for now)
	hostname, _ := os.Hostname()
	a.identity = hostname // In real mTLS, identity is extracted from cert Subject

	// Establish Bi-directional stream
	stream, err := a.client.Connect(context.Background())
	if err != nil {
		return fmt.Errorf("failed to connect stream: %v", err)
	}
	a.stream = stream

	return nil
}

func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.shutdownCh:
			return
		case <-ticker.C:
			a.sendHeartbeat()
		}
	}
}

func (a *Agent) sendHeartbeat() {
	h, _ := host.Info()
	addrs, _ := psnet.Interfaces()
	var ip []string
	if len(addrs) > 0 {
		// Simplified IP picking
		for _, addr := range addrs {
			if len(addr.Addrs) > 0 {
				ip = append(ip, addr.Addrs[0].Addr)
			}
		}
	}

	// In the new proto, everything is wrapped in AgentMessage
	msg := &pb.AgentMessage{
		Payload: &pb.AgentMessage_Heartbeat{
			Heartbeat: &pb.Heartbeat{
				Status: "active",
				Stats:  &pb.SystemStats{
					// We will fill real stats in metricsLoop, but heartbeat can carry them too if needed
					// For now, heartbeat just says "I'm alive"
				},
			},
		},
	}

	// Let's send HostInfo first if this is the first heartbeat (simplified: send every time for MVP)
	hostMsg := &pb.AgentMessage{
		Payload: &pb.AgentMessage_HostInfo{
			HostInfo: &pb.HostInfo{
				Hostname:        h.Hostname,
				Os:              h.OS,
				Platform:        h.Platform,
				PlatformFamily:  h.PlatformFamily,
				PlatformVersion: h.PlatformVersion,
				KernelVersion:   h.KernelVersion,
				Arch:            h.KernelArch,
				IpAddresses:     ip,
				AgentVersion:    "0.1.0",
				UptimeSeconds:   h.Uptime,
			},
		},
	}

	if err := a.stream.Send(hostMsg); err != nil {
		a.logger.Error("Failed to send HostInfo", zap.Error(err))
		return
	}

	if err := a.stream.Send(msg); err != nil {
		a.logger.Error("Heartbeat failed", zap.Error(err))
	} else {
		a.logger.Info("Heartbeat sent")
	}
}

func (a *Agent) metricsLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.shutdownCh:
			return
		case <-ticker.C:
			a.collectAndSendMetrics()
		}
	}
}

func (a *Agent) collectAndSendMetrics() {
	v, _ := mem.VirtualMemory()
	c, _ := cpu.Percent(0, false)
	d, _ := disk.Usage("/")

	cpuVal := 0.0
	if len(c) > 0 {
		cpuVal = c[0]
	}

	// In the new proto, metrics are part of Heartbeat -> SystemStats
	// So we send a Heartbeat message with stats populated
	msg := &pb.AgentMessage{
		Payload: &pb.AgentMessage_Heartbeat{
			Heartbeat: &pb.Heartbeat{
				Status: "active",
				Stats: &pb.SystemStats{
					CpuUsagePercent:    cpuVal,
					MemoryUsagePercent: v.UsedPercent,
					DiskUsagePercent:   d.UsedPercent,
				},
			},
		},
	}

	if err := a.stream.Send(msg); err != nil {
		a.logger.Error("Metrics push failed", zap.Error(err))
	} else {
		a.logger.Info("Metrics pushed", zap.Float64("cpu", cpuVal), zap.Float64("mem", v.UsedPercent))
	}
}

func (a *Agent) taskLoop() {
	// Listen for incoming messages from the stream
	go func() {
		for {
			select {
			case <-a.shutdownCh:
				return
			default:
				msg, err := a.stream.Recv()
				if err != nil {
					a.logger.Error("Stream recv error", zap.Error(err))
					// Reconnect logic needed here
					return
				}

				switch payload := msg.Payload.(type) {
				case *pb.ServerMessage_Task:
					a.handleTask(payload.Task)
				default:
					a.logger.Warn("Unknown message type")
				}
			}
		}
	}()
}

func (a *Agent) handleTask(task *pb.Task) {
	a.logger.Info("Received task", zap.String("command", task.Command))

	// Execute Task
	// Proto definition for Task is very simple: just "Command" string.
	// Real implementation would parse this command or have structured fields.
	// For MVP, we'll assume command is a simple string like "ping" or "collect_metrics"

	err := a.executeTask(task.Command)
	status := "completed"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}

	// Send Response
	resp := &pb.AgentMessage{
		Payload: &pb.AgentMessage_TaskResponse{
			TaskResponse: &pb.TaskResponse{
				Result: fmt.Sprintf("%s: %s", status, errMsg),
			},
		},
	}
	a.stream.Send(resp)
}

func (a *Agent) executeTask(command string) error {
	switch command {
	case "ping":
		return nil
	case "collect_metrics":
		a.collectAndSendMetrics()
		return nil
	case "noop":
		time.Sleep(1 * time.Second)
		return nil
	default:
		return fmt.Errorf("unknown task command: %s", command)
	}
}
