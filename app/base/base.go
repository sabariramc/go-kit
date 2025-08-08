// package base implements a lightweight abstract base framework for a microservice application.
package base

import (
	"context"
	"sync"
	"time"

	"github.com/sabariramc/go-kit/log"
)

// Base represents a basic application structure with configuration, logging, status check, health check, and shutdown functionality.
type Base struct {
	log           *log.Logger       // Logger instance for the application.
	shutdownHooks []ShutdownHook    // List of shutdown hooks to be executed during application shutdown.
	healthHooks   []HealthCheckHook // List of health check hooks.
	shutdownWg    sync.WaitGroup    // WaitGroup for synchronizing shutdown.
}

func New(option ...Option) *Base {
	config := NewDefaultConfig()
	for _, opt := range option {
		opt(config)
	}
	b := &Base{
		shutdownHooks: make([]ShutdownHook, 0, 10),
		healthHooks:   make([]HealthCheckHook, 0, 10),
		log:           config.Log,
	}
	zone, _ := time.Now().Zone()
	b.log.Info(context.TODO()).Msgf("Timezone %v", zone)
	b.shutdownWg.Add(1)
	return b
}
