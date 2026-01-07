package config

import "errors"

var (
	ErrInvalidEnv      = errors.New("invalid_env")
	ErrInvalidPort     = errors.New("invalid_port")
	ErrInvalidShutdown = errors.New("invalid_shutdown_timeout")
)
