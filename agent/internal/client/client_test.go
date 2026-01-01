package client

import (
	"context"
	"testing"
	"time"

	"pulse/agent/internal/config"
	"pulse/agent/internal/storage"
)

func TestUnauthorizedStopsNeedsAttention(t *testing.T) {
	cfg := config.Config{
		GRPCAddress:               "localhost:50051",
		CAFile:                    "testdata/ca.pem",
		CertPath:                  "testdata/client.crt",
		KeyPath:                   "testdata/client.key",
		StatePath:                 "testdata/state.json",
		MaxReconnectAttempts:      1,
		ReconnectBaseDelaySeconds: 1,
		HeartbeatIntervalSeconds:  1,
	}
	r := Runner{Cfg: cfg, Store: storage.CertStore{CertPath: cfg.CertPath, KeyPath: cfg.KeyPath}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { time.Sleep(100 * time.Millisecond); cancel() }()
	err := r.Run(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
}
