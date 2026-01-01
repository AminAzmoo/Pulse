package comm

import (
    "testing"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func TestServerRejectsIdentitySetsNeedsAttention(t *testing.T) {
    err := status.Error(codes.PermissionDenied, "unauthorized")
    s := MapErrorToStatus(err)
    if s != "needs_attention" {
        t.Fatalf("unexpected status %s", s)
    }
}

func TestRevokeStopsReconnect(t *testing.T) {
    err := status.Error(codes.PermissionDenied, "revoked")
    if ShouldReconnect(err) {
        t.Fatalf("expected no reconnect")
    }
}

