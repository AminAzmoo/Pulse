package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	EnvConfigFile           = "PULSE_AGENT_CONFIG_FILE"
	EnvBootstrapURL         = "PULSE_AGENT_BOOTSTRAP_URL"
	EnvInstallToken         = "PULSE_AGENT_INSTALL_TOKEN"
	EnvTokenFile            = "PULSE_AGENT_TOKEN_FILE"
	EnvServerAddress        = "PULSE_AGENT_SERVER_ADDR"
	EnvCACertPath           = "PULSE_AGENT_CA_CERT_PATH"
	EnvCertDir              = "PULSE_AGENT_CERT_DIR"
	EnvCertFile             = "PULSE_AGENT_CERT_FILE"
	EnvKeyFile              = "PULSE_AGENT_KEY_FILE"
	EnvStateFile            = "PULSE_AGENT_STATE_FILE"
	EnvHeartbeatIntervalSec = "PULSE_AGENT_HEARTBEAT_SEC"
	EnvReconnectMaxAttempts = "PULSE_AGENT_RECONNECT_MAX"
	EnvBackoffBaseMs        = "PULSE_AGENT_BACKOFF_BASE_MS"
	EnvBackoffMaxMs         = "PULSE_AGENT_BACKOFF_MAX_MS"
	EnvBackoffJitterRatio   = "PULSE_AGENT_BACKOFF_JITTER"
	EnvRenewBeforeHours     = "PULSE_AGENT_RENEW_BEFORE_H"
	EnvHeavyCapacity        = "PULSE_AGENT_HEAVY_CAPACITY"
	EnvLightCapacity        = "PULSE_AGENT_LIGHT_CAPACITY"
)

const (
	DefaultHeartbeatInterval    = 30 * time.Second
	DefaultReconnectMaxAttempts = 10
	DefaultBackoffBase          = 1 * time.Second
	DefaultBackoffMax           = 60 * time.Second
	DefaultBackoffJitter        = 0.2
	DefaultRenewBefore          = 72 * time.Hour
	DefaultCertDirLinux         = "/var/lib/pulse-agent"
	DefaultCertFileName         = "client.crt"
	DefaultKeyFileName          = "client.key"
	DefaultCACertFileName       = "ca.crt"
	DefaultStateFileName        = "state.json"
)

type Config struct {
	// Bootstrap and control plane
	BootstrapURL    string `json:"bootstrap_url"`
	ControlPlaneURL string `json:"control_plane_url"`
	InstallToken    string `json:"install_token"`
	BootstrapToken  string `json:"bootstrap_token"`
	TokenFile       string `json:"token_file"`

	// gRPC connectivity
	GRPCAddress   string `json:"grpc_address"`
	ServerAddress string `json:"server_address"`
	CAFile        string `json:"ca_file"`

	// Cert storage
	CertDir    string `json:"cert_dir"`
	CertFile   string `json:"cert_file"`
	KeyFile    string `json:"key_file"`
	CACertPath string `json:"ca_cert_path"`
	CertPath   string `json:"cert_path"`
	KeyPath    string `json:"key_path"`

	// Local state
	StateFile string `json:"state_file"`
	StatePath string `json:"state_path"`

	// Timing and reliability
	HeartbeatInterval         time.Duration `json:"heartbeat_interval"`
	HeartbeatIntervalSeconds  int           `json:"heartbeat_interval_seconds"`
	ReconnectMaxAttempts      int           `json:"reconnect_max_attempts"`
	MaxReconnectAttempts      int           `json:"max_reconnect_attempts"`
	ReconnectBaseDelaySeconds int           `json:"reconnect_base_delay_seconds"`
	BackoffBase               time.Duration `json:"backoff_base"`
	BackoffMax                time.Duration `json:"backoff_max"`
	BackoffJitter             float64       `json:"backoff_jitter"`
	RenewBefore               time.Duration `json:"renew_before"`

	// Capacity
	HeavyCapacity         int `json:"heavy_capacity"`
	LightCapacity         int `json:"light_capacity"`
	HeavyConcurrencySlots int `json:"heavy_concurrency_slots"`
	LightConcurrencySlots int `json:"light_concurrency_slots"`

	// Security
	AllowedCommands []string `json:"allowed_commands"`
}

func getenv(key string) string {
	return os.Getenv(key)
}

func atoiEnv(key string, def int) int {
	v := getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func durationEnvMs(key string, def time.Duration) time.Duration {
	n := atoiEnv(key, int(def/time.Millisecond))
	return time.Duration(n) * time.Millisecond
}

func durationEnvSec(key string, def time.Duration) time.Duration {
	n := atoiEnv(key, int(def/time.Second))
	return time.Duration(n) * time.Second
}

func floatEnv(key string, def float64) float64 {
	v := getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

func durationEnvHours(key string, def time.Duration) time.Duration {
	n := atoiEnv(key, int(def/time.Hour))
	return time.Duration(n) * time.Hour
}

func defaults() Config {
	dir := DefaultCertDirLinux
	cert := filepath.Join(dir, DefaultCertFileName)
	key := filepath.Join(dir, DefaultKeyFileName)
	ca := filepath.Join(dir, DefaultCACertFileName)
	state := filepath.Join(dir, DefaultStateFileName)
	return Config{
		HeartbeatInterval:     DefaultHeartbeatInterval,
		ReconnectMaxAttempts:  DefaultReconnectMaxAttempts,
		MaxReconnectAttempts:  DefaultReconnectMaxAttempts,
		BackoffBase:           DefaultBackoffBase,
		BackoffMax:            DefaultBackoffMax,
		BackoffJitter:         DefaultBackoffJitter,
		RenewBefore:           DefaultRenewBefore,
		CertDir:               dir,
		CertFile:              cert,
		CertPath:              cert,
		KeyFile:               key,
		KeyPath:               key,
		CACertPath:            ca,
		StateFile:             state,
		StatePath:             state,
		GRPCAddress:           "",
		ServerAddress:         "",
		CAFile:                ca,
		HeavyCapacity:         0,
		LightCapacity:         0,
		HeavyConcurrencySlots: 1,
		LightConcurrencySlots: 1,
	}
}

func Load() (Config, error) {
	cfg := defaults()
	path := getenv(EnvConfigFile)
	if path != "" {
		b, err := os.ReadFile(path)
		if err == nil {
			_ = json.Unmarshal(b, &cfg)
		}
	}
	if v := getenv(EnvBootstrapURL); v != "" {
		cfg.BootstrapURL = v
	}
	if v := getenv(EnvInstallToken); v != "" {
		cfg.InstallToken = v
	}
	if v := getenv(EnvTokenFile); v != "" {
		cfg.TokenFile = v
	}
	if v := getenv(EnvServerAddress); v != "" {
		cfg.GRPCAddress = v
		cfg.ServerAddress = v
	}
	if v := getenv(EnvCACertPath); v != "" {
		cfg.CACertPath = v
		cfg.CAFile = v
	}
	if v := getenv(EnvCertDir); v != "" {
		cfg.CertDir = v
	}
	if v := getenv(EnvCertFile); v != "" {
		cfg.CertFile = v
		cfg.CertPath = v
	}
	if v := getenv(EnvKeyFile); v != "" {
		cfg.KeyFile = v
		cfg.KeyPath = v
	}
	if v := getenv(EnvStateFile); v != "" {
		cfg.StateFile = v
		cfg.StatePath = v
	}
	cfg.HeartbeatInterval = durationEnvSec(EnvHeartbeatIntervalSec, cfg.HeartbeatInterval)
	cfg.ReconnectMaxAttempts = atoiEnv(EnvReconnectMaxAttempts, cfg.ReconnectMaxAttempts)
	cfg.MaxReconnectAttempts = cfg.ReconnectMaxAttempts
	cfg.BackoffBase = durationEnvMs(EnvBackoffBaseMs, cfg.BackoffBase)
	cfg.BackoffMax = durationEnvMs(EnvBackoffMaxMs, cfg.BackoffMax)
	cfg.BackoffJitter = floatEnv(EnvBackoffJitterRatio, cfg.BackoffJitter)
	cfg.RenewBefore = durationEnvHours(EnvRenewBeforeHours, cfg.RenewBefore)
	cfg.HeavyCapacity = atoiEnv(EnvHeavyCapacity, cfg.HeavyCapacity)
	cfg.LightCapacity = atoiEnv(EnvLightCapacity, cfg.LightCapacity)
	if cfg.HeartbeatIntervalSeconds == 0 {
		cfg.HeartbeatIntervalSeconds = int(cfg.HeartbeatInterval / time.Second)
	}
	if cfg.ReconnectBaseDelaySeconds == 0 {
		cfg.ReconnectBaseDelaySeconds = int(cfg.BackoffBase / time.Second)
	}
	return cfg, nil
}

func Validate(cfg Config) error {
	if cfg.GRPCAddress == "" {
		return errors.New("missing server address")
	}
	if cfg.CertPath == "" || cfg.KeyPath == "" || (cfg.CACertPath == "" && cfg.CAFile == "") {
		return errors.New("missing certificate paths")
	}
	certExists := fileExists(cfg.CertPath)
	keyExists := fileExists(cfg.KeyPath)
	if !certExists || !keyExists {
		if cfg.BootstrapToken == "" && cfg.InstallToken == "" && cfg.TokenFile == "" && cfg.BootstrapURL == "" && cfg.ControlPlaneURL == "" {
			return errors.New("missing bootstrap configuration and certificates")
		}
	}
	return nil
}

func fileExists(p string) bool {
	if p == "" {
		return false
	}
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
