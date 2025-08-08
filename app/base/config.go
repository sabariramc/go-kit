package base

import (
	"github.com/sabariramc/go-kit/log"
)

// Config holds the configuration for the base app.
type Config struct {
	Log *log.Logger
}

// Option represents a function that applies a configuration option to Config.
type Option func(*Config)

func NewDefaultConfig() *Config {
	return &Config{
		Log: log.New("server-base"),
	}
}

// WithLog sets the Log field of Config.
func WithLog(log *log.Logger) Option {
	return func(c *Config) {
		c.Log = log
	}
}
