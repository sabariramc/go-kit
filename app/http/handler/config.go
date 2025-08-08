package handler

import (
	"fmt"

	"github.com/julienschmidt/httprouter"
	"github.com/sabariramc/go-kit/app/http/route"
)

type Config struct {
	Router *httprouter.Router
}

type Option func(*Config) error

func NewConfig(opt ...Option) (*Config, error) {
	router := httprouter.New()
	router.NotFound = route.NotFound()
	router.MethodNotAllowed = route.MethodNotAllowed()
	cfg := &Config{
		Router: router,
	}
	for _, o := range opt {
		err := o(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}
	err := Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

func Validate(c *Config) error {
	return nil
}
