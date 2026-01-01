package grpc

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"pulse/backend/internal/ca"
	"pulse/backend/internal/db/models"
	"pulse/backend/internal/services"
	"pulse/backend/pkg/external/grpc/pb"
	"pulse/backend/pkg/external/logging"
)

const (
	HeartbeatInterval     = 30 * time.Second
	DenylistCheckInterval = 5 * time.Minute
	MaxStreamIdle         = 10 * time.Minute
)

type IdentityProvider interface {
	IsAllowed(identity string) (bool, error)
	Subscribe(identity string) chan struct{}
	Unsubscribe(identity string, ch chan struct{})
	UpdateHostInfo(identity string, info *models.Node) error
	RecordMetrics(identity string, metrics *models.Metric) error
}

type AgentService struct {
	pb.UnimplementedAgentServiceServer
	IdentitySvc IdentityProvider
	TaskSvc     *services.TaskService
	CASvc       *ca.CAService
}

func (s *AgentService) Connect(stream pb.AgentService_ConnectServer) error {
	// Get client identity from mTLS cert
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no peer info")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return status.Error(codes.Unauthenticated, "unexpected auth info")
	}

	certs := tlsInfo.State.PeerCertificates
	if len(certs) == 0 {
		return status.Error(codes.Unauthenticated, "no client cert")
	}

	identity := certs[0].Subject.CommonName

	// Check allowlist
	allowed, err := s.IdentitySvc.IsAllowed(identity)
	if err != nil {
		logging.Logger.Error("Allowlist check failed", zap.Error(err), zap.String("identity", identity))
		return status.Error(codes.Internal, "internal error")
	}

	if !allowed {
		logging.Audit("Unauthorized connection attempt", map[string]interface{}{"identity": identity})
		return status.Error(codes.PermissionDenied, "unauthorized")
	}

	// Subscribe to revocation notifications
	revokeChan := s.IdentitySvc.Subscribe(identity)
	defer s.IdentitySvc.Unsubscribe(identity, revokeChan)

	logging.Logger.Info("Agent connected", zap.String("identity", identity))

	// Stream loop
	lastHeartbeat := time.Now()
	heavyAvailable := 0
	lightAvailable := 0

	type recvMsg struct {
		msg *pb.AgentMessage
		err error
	}
	recvChan := make(chan recvMsg, 1)

	go func() {
		for {
			msg, err := stream.Recv()
			select {
			case recvChan <- recvMsg{msg: msg, err: err}:
			case <-stream.Context().Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-revokeChan:
			logging.Logger.Info("Revocation detected, terminating stream", zap.String("identity", identity))
			return status.Error(codes.PermissionDenied, "revoked")
		case rm := <-recvChan:
			if rm.err == io.EOF {
				return nil
			}
			if rm.err != nil {
				return rm.err
			}

			if t := rm.msg.GetHostInfo(); t != nil {
				// Process host info
				node := &models.Node{
					Hostname:        t.Hostname,
					OS:              t.Os,
					Platform:        t.Platform,
					PlatformFamily:  t.PlatformFamily,
					PlatformVersion: t.PlatformVersion,
					KernelVersion:   t.KernelVersion,
					Arch:            t.Arch,
					AgentVersion:    t.AgentVersion,
					UptimeSeconds:   t.UptimeSeconds,
					IPAddresses:     strings.Join(t.IpAddresses, ","),
				}
				if err := s.IdentitySvc.UpdateHostInfo(identity, node); err != nil {
					logging.Logger.Error("Failed to update host info", zap.Error(err))
				}
			} else if h := rm.msg.GetHeartbeat(); h != nil {
				// Process heartbeat and stats
				if h.Stats != nil {
					m := &models.Metric{
						CPU:    h.Stats.CpuUsagePercent,
						Memory: h.Stats.MemoryUsagePercent,
						Disk:   h.Stats.DiskUsagePercent,
					}
					// Fire and forget metrics recording to avoid blocking
					go func() {
						if err := s.IdentitySvc.RecordMetrics(identity, m); err != nil {
							logging.Logger.Warn("Failed to record metrics", zap.Error(err))
						}
					}()
				}
				lastHeartbeat = time.Now()
				// deliver tasks if capacity known
				if heavyAvailable > 0 || lightAvailable > 0 {
					envelopes, derr := s.TaskSvc.NextTasksFor(identity, heavyAvailable, lightAvailable)
					if derr == nil {
						for _, env := range envelopes {
							cmdBytes, _ := json.Marshal(env)
							if err := stream.Send(&pb.ServerMessage{Payload: &pb.ServerMessage_Task{Task: &pb.Task{
								Command: string(cmdBytes),
							}}}); err != nil {
								return err
							}
							_ = s.TaskSvc.MarkInProgress(env.TaskID)
						}
					}
				}
			} else if resp := rm.msg.GetTaskResponse(); resp != nil {
				var update services.TaskStatusUpdate
				if err := json.Unmarshal([]byte(resp.Result), &update); err == nil {
					if update.Kind == services.UpdateKindCapacity {
						heavyAvailable = update.HeavyAvailable
						lightAvailable = update.LightAvailable
					} else {
						_ = s.TaskSvc.HandleStatusUpdate(update)
					}
				}
			}
		case <-time.After(MaxStreamIdle):
			if time.Since(lastHeartbeat) > MaxStreamIdle {
				return status.Error(codes.DeadlineExceeded, "stream idle")
			}
		}
	}
}

func (s *AgentService) RenewCertificate(ctx context.Context, req *pb.RenewCertificateRequest) (*pb.RenewCertificateResponse, error) {
	// Authenticate via mTLS
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no peer info")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unexpected auth info")
	}

	certs := tlsInfo.State.PeerCertificates
	if len(certs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no client cert")
	}

	identity := certs[0].Subject.CommonName

	// Check allowlist
	allowed, err := s.IdentitySvc.IsAllowed(identity)
	if err != nil {
		logging.Logger.Error("Allowlist check failed", zap.Error(err), zap.String("identity", identity))
		return nil, status.Error(codes.Internal, "internal error")
	}

	if !allowed {
		logging.Audit("Unauthorized renewal attempt", map[string]interface{}{"identity": identity})
		return nil, status.Error(codes.PermissionDenied, "unauthorized")
	}

	// Issue new certificate
	// Note: We currently ignore req.Csr because our CA service generates the keypair.
	certPEM, keyPEM, err := s.CASvc.IssueCertificate(identity)
	if err != nil {
		logging.Logger.Error("Failed to issue certificate", zap.Error(err), zap.String("identity", identity))
		return nil, status.Error(codes.Internal, "failed to issue certificate")
	}

	return &pb.RenewCertificateResponse{
		Certificate:   certPEM + "\n" + keyPEM,
		CaCertificate: "", // Client already has CA cert
	}, nil
}
