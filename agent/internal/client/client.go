package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"

	"pulse/agent/internal/collector"
	"pulse/agent/internal/config"
	"pulse/agent/internal/executor"
	"pulse/agent/internal/state"
	"pulse/agent/internal/storage"
	"pulse/agent/internal/tasks"
	pb "pulse/backend/pkg/external/grpc/pb"
)

type Runner struct {
	Cfg   config.Config
	Store storage.CertStore
}

const (
	stateRunning        = "running"
	stateNeedsAttention = "needs_attention"
	stateRevoked        = "revoked"
)

func (r Runner) Run(ctx context.Context) error {
	certPEM, keyPEM, err := r.Store.Load()
	if err != nil {
		return errors.New("bootstrap required")
	}
	ca, err := loadCA(r.Cfg.CAFile)
	if err != nil {
		return err
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return err
	}
	pool := x509.NewCertPool()
	_ = pool.AppendCertsFromPEM(ca)
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: pool, Certificates: []tls.Certificate{cert}}

	attempts := 0
	for attempts < r.Cfg.MaxReconnectAttempts {
		attempts++
		conn, err := grpc.NewClient(r.Cfg.GRPCAddress, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
		if err != nil {
			delay := time.Duration(r.Cfg.ReconnectBaseDelaySeconds*attempts) * time.Second
			zap.L().Warn("connect failed", zap.Error(err), zap.Int("attempt", attempts))
			time.Sleep(delay)
			continue
		}
		c := pb.NewAgentServiceClient(conn)

		if r.maybeRenew(ctx, c, cert) {
			_ = conn.Close()
			certPEM, keyPEM, err := r.Store.Load()
			if err != nil {
				return err
			}
			cert, err = tls.X509KeyPair(certPEM, keyPEM)
			if err != nil {
				return err
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
			attempts = 0
			continue
		}

		stream, err := c.Connect(ctx)
		if err != nil {
			_ = conn.Close()
			if st, ok := status.FromError(err); ok && st.Code() == codes.PermissionDenied {
				_ = state.Save(r.Cfg.StatePath, stateNeedsAttention, map[string]string{"reason": "unauthorized"})
				return errors.New("unauthorized")
			}
			delay := time.Duration(r.Cfg.ReconnectBaseDelaySeconds*attempts) * time.Second
			time.Sleep(delay)
			continue
		}

		renewalCtx, cancelRenewal := context.WithCancel(ctx)
		go func() {
			defer cancelRenewal()
			ticker := time.NewTicker(12 * time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-renewalCtx.Done():
					return
				case <-ticker.C:
					leaf, _ := x509.ParseCertificate(cert.Certificate[0])
					if time.Until(leaf.NotAfter) < 30*24*time.Hour {
						_ = conn.Close()
						return
					}
				}
			}
		}()

		hb := time.NewTicker(time.Duration(r.Cfg.HeartbeatIntervalSeconds) * time.Second)
		_ = state.Save(r.Cfg.StatePath, stateRunning, map[string]string{"attempt": strconvI(attempts)})
		reg := tasks.NewRegistry(r.Cfg.AllowedCommands)
		exec := executor.NewExecutor(reg, r.Cfg.HeavyConcurrencySlots, r.Cfg.LightConcurrencySlots, r.Cfg.StatePath)
		runErr := runStream(ctx, stream, hb, exec)
		cancelRenewal()
		hb.Stop()
		_ = conn.Close()
		if runErr != nil {
			if st, ok := status.FromError(runErr); ok && st.Code() == codes.PermissionDenied {
				_ = state.Save(r.Cfg.StatePath, stateRevoked, map[string]string{"reason": "revoked"})
				return errors.New("revoked")
			}
		}
		delay := time.Duration(r.Cfg.ReconnectBaseDelaySeconds*attempts) * time.Second
		time.Sleep(delay)
	}
	_ = state.Save(r.Cfg.StatePath, stateNeedsAttention, map[string]string{"reason": "max_retries"})
	return errors.New("max retries reached")
}

type AgentStream interface {
	Send(*pb.AgentMessage) error
	Recv() (*pb.ServerMessage, error)
}

func runStream(ctx context.Context, stream AgentStream, hb *time.Ticker, exec *executor.Executor) error {
	// announce capacity to backend
	heavy, light := exec.Capacity()
	capMsg := map[string]interface{}{"kind": "capacity", "heavy_available": heavy, "light_available": light}
	capJSON, _ := jsonMarshal(capMsg)
	_ = stream.Send(&pb.AgentMessage{Payload: &pb.AgentMessage_TaskResponse{TaskResponse: &pb.TaskResponse{Result: capJSON}}})

	// Announce Host Info
	hostInfo := collector.GetHostInfo()
	_ = stream.Send(&pb.AgentMessage{Payload: &pb.AgentMessage_HostInfo{HostInfo: hostInfo}})

	errChan := make(chan error, 1)

	// Receiver goroutine
	go func() {
		for {
			srvMsg, err := stream.Recv()
			if err != nil {
				errChan <- err
				return
			}
			if t := srvMsg.GetTask(); t != nil {
				var env executor.Envelope
				if err := jsonUnmarshal([]byte(t.GetCommand()), &env); err == nil {
					reporter := &streamReporter{stream: stream}
					go exec.Execute(env, reporter)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		case <-hb.C:
			msg := &pb.AgentMessage{
				Payload: &pb.AgentMessage_Heartbeat{
					Heartbeat: &pb.Heartbeat{
						Status: "ok",
						Stats:  collector.Collect(),
					},
				},
			}
			if err := stream.Send(msg); err != nil {
				return err
			}
		}
	}
}

type streamReporter struct{ stream AgentStream }

func (sr *streamReporter) ReportStatus(taskID uint, status string, errMsg string) error {
	msg := map[string]interface{}{"kind": "status", "task_id": taskID, "status": status}
	if errMsg != "" {
		msg["error"] = errMsg
	}
	s, _ := jsonMarshal(msg)
	return sr.stream.Send(&pb.AgentMessage{Payload: &pb.AgentMessage_TaskResponse{TaskResponse: &pb.TaskResponse{Result: s}}})
}

func jsonMarshal(v interface{}) (string, error)   { b, e := json.Marshal(v); return string(b), e }
func jsonUnmarshal(b []byte, v interface{}) error { return json.Unmarshal(b, v) }

func loadCA(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func strconvI(i int) string { return strconv.Itoa(i) }

func (r Runner) maybeRenew(ctx context.Context, client pb.AgentServiceClient, tlsCert tls.Certificate) bool {
	leaf, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return false
	}

	// Renew if expires in < 30 days
	if time.Until(leaf.NotAfter) > 30*24*time.Hour {
		return false
	}

	zap.L().Info("Certificate expiring soon, attempting renewal", zap.Time("expires", leaf.NotAfter))

	resp, err := client.RenewCertificate(ctx, &pb.RenewCertificateRequest{})
	if err != nil {
		zap.L().Error("Renewal failed", zap.Error(err))
		return false
	}

	certPEM, keyPEM, err := parseCertAndKey([]byte(resp.Certificate))
	if err != nil {
		zap.L().Error("Failed to parse renewed cert", zap.Error(err))
		return false
	}

	if err := r.Store.Save(certPEM, keyPEM); err != nil {
		zap.L().Error("Failed to save renewed cert", zap.Error(err))
		return false
	}

	zap.L().Info("Certificate renewed successfully")
	return true
}

func parseCertAndKey(blob []byte) ([]byte, []byte, error) {
	var certPEM, keyPEM []byte
	rest := blob
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		b := pem.EncodeToMemory(block)
		if block.Type == "CERTIFICATE" {
			certPEM = append(certPEM, b...)
		} else if block.Type == "RSA PRIVATE KEY" || block.Type == "EC PRIVATE KEY" || block.Type == "PRIVATE KEY" {
			keyPEM = b
		}
	}
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return nil, nil, errors.New("incomplete cert/key pair")
	}
	return certPEM, keyPEM, nil
}
