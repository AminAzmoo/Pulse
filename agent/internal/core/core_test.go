package core

import (
    "testing"
    "pulse/agent/internal/config"
)

func TestMissingTokenAndCertFailsSafe(t *testing.T) {
    cfg := config.Config{
        ServerAddress: "localhost:50051",
        CACertPath: "",
        CertFile: "",
        KeyFile: "",
        BootstrapURL: "",
        InstallToken: "",
        TokenFile: "",
    }
    err := config.Validate(cfg)
    if err == nil {
        t.Fatalf("expected error")
    }
}

