package http

import (
	"fmt"
	"net/http"

	"github.com/sabariramc/go-kit/app/base"
	"github.com/sabariramc/go-kit/log"
)

type Config struct {
	Base    *base.Base
	Log     *log.Logger
	Server  *http.Server
	Handler http.Handler
}

func NewConfig(opt ...Option) (*Config, error) {
	cfg := &Config{
		Base: base.New(),
		Log:  log.New("HTTPServer"),
	}
	for _, o := range opt {
		err := o(cfg)
		if err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}
	if cfg.Server == nil {
		cfg.Server = &http.Server{
			Addr:    ":8080",
			Handler: cfg.Handler,
		}
	}
	err := Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

func Validate(c *Config) error {
	if c.Base == nil {
		return fmt.Errorf("base is required")
	}
	if c.Log == nil {
		return fmt.Errorf("log is required")
	}
	if c.Handler == nil {
		return fmt.Errorf("handler is required")
	}
	return nil
}

type Option func(*Config) error

func WithHandler(handler http.Handler) Option {
	return func(c *Config) error {
		c.Handler = handler
		return nil
	}
}
