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

	if _, err := net.LookupPort("tcp", c.Port); err != nil {
		return fmt.Errorf("port=%s: %w", c.Port, ErrInvalidPort)
	}

	if c.ShutdownSecs <= 0 || c.ShutdownSecs > 120 {
		return fmt.Errorf("shutdown_secs=%d: %w", c.ShutdownSecs, ErrInvalidShutdown)
	}

	return nil
}
