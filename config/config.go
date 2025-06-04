package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the operator
type Config struct {
	Database struct {
		User     string
		Password string
		Host     string
		Port     string
		Name     string
		Pool     struct {
			MaxOpenConns    int
			MaxIdleConns    int
			ConnMaxLifetime time.Duration
		}
	}
	Kubernetes struct {
		ConfigPath string
		Namespace  string
	}
	Operator struct {
		PollInterval time.Duration
		LogLevel     string
	}
}

// NewDefaultConfig returns a Config with default values
func NewDefaultConfig() *Config {
	cfg := &Config{}
	
	// Database defaults
	cfg.Database.Port = "3306"
	cfg.Database.Pool.MaxOpenConns = 10
	cfg.Database.Pool.MaxIdleConns = 5
	cfg.Database.Pool.ConnMaxLifetime = time.Hour

	// Kubernetes defaults
	cfg.Kubernetes.ConfigPath = "/etc/kubernetes/admin.conf"
	cfg.Kubernetes.Namespace = "uusec"

	// Operator defaults
	cfg.Operator.PollInterval = 5 * time.Second
	cfg.Operator.LogLevel = "info"

	return cfg
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
} 