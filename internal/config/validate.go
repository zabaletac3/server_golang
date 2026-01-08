package config

import (
	"fmt"
	"net"
)

var allowedEnvs = map[string]struct{}{
	"development": {},
	"staging":     {},
	"production":  {},
}

func (c *Config) Validate() error {
	if _, ok := allowedEnvs[c.Env]; !ok {
		return fmt.Errorf("env=%s: %w", c.Env, ErrInvalidEnv)
	}

	if c.Port == "" {
		return fmt.Errorf("port is required: %w", ErrInvalidPort)
	}

	if _, err := net.LookupPort("tcp", c.Port); err != nil {
		return fmt.Errorf("port=%s: %w", c.Port, ErrInvalidPort)
	}

	if c.ShutdownSecs <= 0 || c.ShutdownSecs > 120 {
		return fmt.Errorf("shutdown_secs=%d: %w", c.ShutdownSecs, ErrInvalidShutdown)
	}

	if c.ReadHeaderTimeoutSecs <= 0 {
		return fmt.Errorf("read_header_timeout invalid")
	}
	if c.ReadTimeoutSecs <= 0 {
		return fmt.Errorf("read_timeout invalid")
	}
	if c.WriteTimeoutSecs <= 0 {
		return fmt.Errorf("write_timeout invalid")
	}
	if c.IdleTimeoutSecs <= 0 {
		return fmt.Errorf("idle_timeout invalid")
	}
	if c.MaxHeaderBytes <= 0 {
		return fmt.Errorf("max_header_bytes invalid")
	}

	return nil
}
