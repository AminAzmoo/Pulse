package client

import (
	"testing"

	pb "pulse/backend/pkg/external/grpc/pb"
)

type benchStream struct{}

func (b *benchStream) Send(*pb.AgentMessage) error      { return nil }
func (b *benchStream) Recv() (*pb.ServerMessage, error) { return nil, nil }

const statusOK = "ok"

func BenchmarkEncodeSendHeartbeat(b *testing.B) {
	s := &benchStream{}
	for i := 0; i < b.N; i++ {
		_ = s.Send(&pb.AgentMessage{Payload: &pb.AgentMessage_Heartbeat{Heartbeat: &pb.Heartbeat{Status: statusOK}}})
	}
}
