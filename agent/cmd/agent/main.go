package main

import (
    "context"
    "os"
    "time"

    "go.uber.org/zap"

    "pulse/agent/internal/bootstrap"
    "pulse/agent/internal/client"
    "pulse/agent/internal/config"
    "pulse/agent/internal/logging"
    "pulse/agent/internal/state"
    "pulse/agent/internal/storage"
)

const bootstrapPath = "v1/agent/bootstrap"

func main() {
    _ = logging.Init()
    cfg, err := config.Load()
    if err != nil {
        zap.L().Fatal("config error", zap.Error(err))
    }

    store := storage.CertStore{CertPath: cfg.CertPath, KeyPath: cfg.KeyPath}
    _, _, loadErr := store.Load()
    if loadErr != nil {
        if cfg.BootstrapToken == "" {
            _ = state.Save(cfg.StatePath, "needs_attention", map[string]string{"reason": "missing_token"})
            zap.L().Fatal("bootstrap token missing")
        }
        caPEM, _ := os.ReadFile(cfg.CAFile)
        cert, key, berr := bootstrap.ExchangeToken(cfg.ControlPlaneURL, bootstrapPath, cfg.BootstrapToken, caPEM)
        if berr != nil {
            _ = state.Save(cfg.StatePath, "needs_attention", map[string]string{"reason": "bootstrap_failed"})
            zap.L().Fatal("bootstrap failed", zap.Error(berr))
        }
        if err := store.Save([]byte(cert), []byte(key)); err != nil {
            _ = state.Save(cfg.StatePath, "needs_attention", map[string]string{"reason": "store_failed"})
            zap.L().Fatal("store cert failed", zap.Error(err))
        }
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    r := client.Runner{Cfg: cfg, Store: store}
    if err := r.Run(ctx); err != nil {
        time.Sleep(1 * time.Second)
        os.Exit(1)
    }
}

