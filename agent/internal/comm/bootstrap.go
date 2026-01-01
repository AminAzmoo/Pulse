package comm

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"pulse/agent/internal/config"
	"pulse/agent/internal/logger"
	"pulse/agent/internal/utils"
	"strings"
	"time"

	"go.uber.org/zap"
)

type BootstrapResponse struct {
	Identity  string `json:"identity"`
	CertPEM   string `json:"cert_pem"`
	KeyPEM    string `json:"key_pem"`
	CACertPEM string `json:"ca_cert_pem"`
}

func readToken(cfg config.Config) (string, error) {
	if cfg.InstallToken != "" {
		return cfg.InstallToken, nil
	}
	if cfg.TokenFile != "" {
		b, err := os.ReadFile(cfg.TokenFile)
		if err != nil {
			return "", err
		}
		t := strings.TrimSpace(string(b))
		if t == "" {
			return "", errors.New("empty token")
		}
		return t, nil
	}
	return "", errors.New("missing token")
}

func Bootstrap(cfg config.Config) (BootstrapResponse, error) {
	var out BootstrapResponse
	token, err := readToken(cfg)
	if err != nil {
		return out, err
	}
	body := map[string]string{"token": token}
	b, _ := json.Marshal(body)
	cli := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", cfg.BootstrapURL, strings.NewReader(string(b)))
	if err != nil {
		return out, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return out, errors.New("bootstrap rejected")
	}
	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(rb, &out); err != nil {
		return out, err
	}
	if out.Identity == "" || out.CertPEM == "" || out.KeyPEM == "" || out.CACertPEM == "" {
		return out, errors.New("invalid bootstrap response")
	}
	if err := utils.EnsureSecureDir(cfg.CertDir, utils.DefaultOSAdapter{}); err != nil {
		return out, err
	}
	if err := os.WriteFile(cfg.CertFile, []byte(out.CertPEM), 0o600); err != nil {
		return out, err
	}
	if err := os.WriteFile(cfg.KeyFile, []byte(out.KeyPEM), 0o600); err != nil {
		return out, err
	}
	if err := os.WriteFile(cfg.CACertPath, []byte(out.CACertPEM), 0o600); err != nil {
		return out, err
	}
	_ = utils.EnsureSecureFile(cfg.CertFile, utils.DefaultOSAdapter{})
	_ = utils.EnsureSecureFile(cfg.KeyFile, utils.DefaultOSAdapter{})
	_ = utils.EnsureSecureFile(cfg.CACertPath, utils.DefaultOSAdapter{})
	if cfg.TokenFile != "" {
		_ = os.Remove(cfg.TokenFile)
	}
	h := sha256.Sum256([]byte(token))
	logger.L.Info("Bootstrap completed", zap.String("identity", out.Identity), zap.String("token_hash", hex.EncodeToString(h[:])))
	return out, nil
}
