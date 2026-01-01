package comm

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "encoding/json"
    "errors"
    "os"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/status"

    "pulse/agent/internal/config"
    "pulse/agent/internal/logger"
    pb "pulse/backend/pkg/external/grpc/pb"
)

type ConnectResult struct {
	Err error
}

func tlsConfigFromFiles(cfg config.Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, err
	}
    caBytes, err := os.ReadFile(cfg.CACertPath)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBytes) {
		return nil, errors.New("invalid ca cert")
	}
	t := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		MinVersion:   tls.VersionTLS13,
	}
	return t, nil
}

func DialClient(cfg config.Config) (*grpc.ClientConn, error) {
    t, err := tlsConfigFromFiles(cfg)
    if err != nil {
        return nil, err
    }
    creds := credentials.NewTLS(t)
    conn, err := grpc.NewClient(cfg.ServerAddress, grpc.WithTransportCredentials(creds))
    if err != nil {
        return nil, err
    }
    return conn, nil
}

func RunStream(ctx context.Context, cfg config.Config) ConnectResult {
	conn, err := DialClient(cfg)
	if err != nil {
		return ConnectResult{Err: err}
	}
	defer conn.Close()
	cli := pb.NewAgentServiceClient(conn)
	stream, err := cli.Connect(ctx)
	if err != nil {
		return ConnectResult{Err: err}
	}
	if logger.L != nil {
		logger.L.Info("Stream connected")
	}
	hb := &pb.AgentMessage{Payload: &pb.AgentMessage_Heartbeat{Heartbeat: &pb.Heartbeat{Status: "starting"}}}
	_ = stream.Send(hb)
	lastHB := time.Now()
	heavy := cfg.HeavyCapacity
	light := cfg.LightCapacity
	for {
		srvMsg, recvErr := stream.Recv()
		if recvErr != nil {
			return ConnectResult{Err: recvErr}
		}
		if srvMsg.GetTask() != nil {
			envJSON := srvMsg.GetTask().GetCommand()
			var payload map[string]interface{}
			_ = json.Unmarshal([]byte(envJSON), &payload)
			_ = stream.Send(&pb.AgentMessage{Payload: &pb.AgentMessage_TaskResponse{TaskResponse: &pb.TaskResponse{Result: string(marshalUpdateStatus(payload))}}})
		}
		if time.Since(lastHB) >= cfg.HeartbeatInterval {
			_ = stream.Send(&pb.AgentMessage{Payload: &pb.AgentMessage_Heartbeat{Heartbeat: &pb.Heartbeat{Status: "ok"}}})
			if heavy > 0 || light > 0 {
				cu := map[string]interface{}{"kind": "capacity", "heavy_available": heavy, "light_available": light}
				b, _ := json.Marshal(cu)
				_ = stream.Send(&pb.AgentMessage{Payload: &pb.AgentMessage_TaskResponse{TaskResponse: &pb.TaskResponse{Result: string(b)}}})
			}
			lastHB = time.Now()
		}
	}
}

func marshalUpdateStatus(payload map[string]interface{}) []byte {
	u := map[string]interface{}{"kind": "status", "status": "in_progress"}
	b, _ := json.Marshal(u)
	return b
}

func MapErrorToStatus(err error) string {
	if err == nil {
		return ""
	}
	st, ok := status.FromError(err)
	if !ok {
		return "offline"
	}
	if st.Code() == codes.PermissionDenied {
		return "needs_attention"
	}
	if st.Code() == codes.Unauthenticated {
		return "needs_attention"
	}
	return "offline"
}

func ShouldReconnect(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return true
	}
	if st.Code() == codes.PermissionDenied {
		return false
	}
	return true
}
