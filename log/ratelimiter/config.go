package ratelimiter

import (
	"fmt"
	"time"
)

type Config struct {
	KeyConfig    KeyConfig
	BlockSize    time.Duration
	WindowSize   time.Duration
	MinBlockSize int
	queueLength  int // Maximum queue length
}

func NewDefaultConfig() *Config {
	return &Config{
		KeyConfig: KeyConfig{
			SizeLimit:        100,
			sizeLimitEnabled: true,
		},
		BlockSize:    10 * time.Second,
		WindowSize:   60 * time.Second,
		MinBlockSize: 100,
	}
}

func ValidateConfig(cfg *Config) error {
	if cfg.KeyConfig.SizeLimit <= 0 {
		return fmt.Errorf("invalid key size limit: %d", cfg.KeyConfig.SizeLimit)
	} else {
		cfg.KeyConfig.sizeLimitEnabled = true
	}
	if cfg.BlockSize <= 0 {
		return fmt.Errorf("invalid block size: %d", cfg.BlockSize)
	}
	if cfg.WindowSize <= 0 {
		return fmt.Errorf("invalid window size: %d", cfg.WindowSize)
	}
	if cfg.MinBlockSize <= 0 {
		return fmt.Errorf("invalid min block size: %d", cfg.MinBlockSize)
	}
	queueLength := int(cfg.WindowSize / cfg.BlockSize)
	if queueLength <= 0 {
		return fmt.Errorf("WindowSize must be greater than BlockSize")
	}
	cfg.queueLength = queueLength
	return nil
}

type Option func(*Config)

func WithKeyExtractor(extractor KeyExtractor) Option {
	return func(cfg *Config) {
		cfg.KeyConfig.Extractor = extractor
	}
}
