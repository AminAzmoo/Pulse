package core

import (
	"context"
	"errors"
	"os"
	"pulse/agent/internal/comm"
	"pulse/agent/internal/config"
	"pulse/agent/internal/logger"
	"pulse/agent/internal/state"
	"pulse/agent/internal/utils"
	"time"

	"google.golang.org/grpc/status"
)

type App struct {
	Cfg config.Config
	SM  *state.Manager
}

func NewApp(cfg config.Config) *App {
	return &App{Cfg: cfg, SM: &state.Manager{Path: cfg.StateFile}}
}

func (a *App) Prepare() error {
	if err := logger.Setup(); err != nil {
		return err
	}
	if err := config.Validate(a.Cfg); err != nil {
		return err
	}
	if !fileExists(a.Cfg.CertFile) || !fileExists(a.Cfg.KeyFile) || !fileExists(a.Cfg.CACertPath) {
		if a.Cfg.BootstrapURL == "" {
			return errors.New("bootstrap not configured")
		}
		_, err := comm.Bootstrap(a.Cfg)
		if err != nil {
			return err
		}
		if err := utils.EnsureSecureFile(a.Cfg.CertFile, utils.DefaultOSAdapter{}); err != nil {
			return err
		}
		if err := utils.EnsureSecureFile(a.Cfg.KeyFile, utils.DefaultOSAdapter{}); err != nil {
			return err
		}
		if err := utils.EnsureSecureFile(a.Cfg.CACertPath, utils.DefaultOSAdapter{}); err != nil {
			return err
		}
	}
	s := state.State{AgentStatus: state.StatusStarting}
	_ = a.SM.Save(s)
	return nil
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	attempt := 0
	for {
		r := comm.RunStream(ctx, a.Cfg)
		if r.Err == nil {
			a.SM.Save(state.State{AgentStatus: state.StatusRunning, LastConnectionTime: time.Now()})
			return nil
		}
		statusStr := comm.MapErrorToStatus(r.Err)
		if statusStr == "needs_attention" {
			a.SM.Save(state.State{AgentStatus: state.StatusNeedsAttention, LastErrorCode: status.Code(r.Err).String()})
			return nil
		}
		if attempt >= a.Cfg.ReconnectMaxAttempts {
			a.SM.Save(state.State{AgentStatus: state.StatusOffline, LastErrorCode: status.Code(r.Err).String()})
			return r.Err
		}
		d := utils.ExponentialBackoff(attempt, a.Cfg.BackoffBase, a.Cfg.BackoffMax, a.Cfg.BackoffJitter)
		time.Sleep(d)
		attempt++
	}
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
