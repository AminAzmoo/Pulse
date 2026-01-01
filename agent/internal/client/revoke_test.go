package client

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"pulse/agent/internal/executor"
	"pulse/agent/internal/tasks"
	pb "pulse/backend/pkg/external/grpc/pb"
)

type fakeStream struct{}

func (f *fakeStream) Send(*pb.AgentMessage) error {
	return status.Error(codes.PermissionDenied, "revoked")
}
func (f *fakeStream) Recv() (*pb.ServerMessage, error) { select {} }
func (f *fakeStream) Header() (metadata.MD, error)     { return metadata.MD{}, nil }
func (f *fakeStream) Trailer() metadata.MD             { return metadata.MD{} }
func (f *fakeStream) CloseSend() error                 { return nil }
func (f *fakeStream) Context() context.Context         { return context.Background() }

func TestRunStreamRevoked(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hb := time.NewTicker(10 * time.Millisecond)
	defer hb.Stop()
	reg := tasks.NewRegistry(nil)
	tmp := t.TempDir() + "/state.json"
	exec := executor.NewExecutor(reg, 1, 1, tmp)
	err := runStream(ctx, &fakeStream{}, hb, exec)
	if err == nil {
		t.Fatal("expected revoked error")
	}
}
