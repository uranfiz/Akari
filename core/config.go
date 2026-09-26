package core

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var Cfg *Config

type Config struct {
	mu sync.RWMutex `yaml:"-"`

	APIID   int32  `yaml:"api_id"`
	APIHash string `yaml:"api_hash"`

	SessionName string `yaml:"session_name"`

	OwnerID    int64  `yaml:"owner_id"`
	OwnerPhone string `yaml:"owner_phone"`

	Lang   string `yaml:"lang"`
	Prefix string `yaml:"prefix"`

	ModulesDir      string   `yaml:"modules_dir"`
	SystemModules   []string `yaml:"system_modules"`
	DisabledModules []string `yaml:"disabled_modules"`

	GitHubRepo   string `yaml:"github_repo"`
	GitHubBranch string `yaml:"github_branch"`
	RepoDir      string `yaml:"repo_dir"`
	AutoUpdate   bool   `yaml:"auto_update"`

	LogLevel string `yaml:"log_level"`

	// Device fingerprint fields (persisted for session stability)
	DeviceModel   string `yaml:"device_model"`
	SystemVersion string `yaml:"system_version"`
	AppVersion    string `yaml:"app_version"`

	path string `yaml:"-"`
}

func DefaultConfig() *Config {
	return &Config{
		SessionName:     "akari",
		Lang:            "ru",
		Prefix:          ".",
		ModulesDir:      "userland",
		SystemModules:   []string{"ping", "restart", "loader", "modules", "terminal", "update"},
		DisabledModules: []string{},
		GitHubRepo:      "uranfiz/Akari",
		GitHubBranch:    "main",
		RepoDir:         ".",
		AutoUpdate:      true,
		LogLevel:        "info",
	}
}

func ConfigPath() string {
	if p := os.Getenv("AKARI_CONFIG_PATH"); p != "" {
		return p
	}
	return "config.yaml"
}

func LoadConfig() (*Config, error) {
	path := ConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.path = path
	cfg.normalize()
	Cfg = cfg
	return cfg, nil
}

func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.path == "" {
		c.path = ConfigPath()
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	tmp := c.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmp, c.path); err != nil {
		return fmt.Errorf("rename temp config: %w", err)
	}
	return nil
}

func IsFirstRun() bool {
	_, err := os.Stat(ConfigPath())
	return errors.Is(err, os.ErrNotExist)
}

func (c *Config) normalize() {
	def := DefaultConfig()

	if c.SessionName == "" {
		c.SessionName = def.SessionName
	}
	if c.Lang == "" {
		c.Lang = def.Lang
	}
	if c.Prefix == "" {
		c.Prefix = def.Prefix
	}
	if c.ModulesDir == "" {
		c.ModulesDir = def.ModulesDir
	}
	if len(c.SystemModules) == 0 {
		c.SystemModules = def.SystemModules
	}
	if c.GitHubRepo == "" {
		c.GitHubRepo = def.GitHubRepo
	}
	if c.GitHubBranch == "" {
		c.GitHubBranch = def.GitHubBranch
	}
	if c.RepoDir == "" {
		c.RepoDir = def.RepoDir
	}
	if c.LogLevel == "" {
		c.LogLevel = def.LogLevel
	}
}

func (c *Config) IsOwner(userID int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.OwnerID != 0 && c.OwnerID == userID
}

func (c *Config) IsSystemModule(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, m := range c.SystemModules {
		if m == name {
			return true
		}
	}
	return false
}
