package base

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

// RegisterHealthCheckHook registers a health check hook to be executed during the health check.
//
// This function appends the provided health check handler to the list of health check hooks in the BaseApp.
func (b *Base) RegisterHealthCheckHook(handler HealthCheckHook) {
	b.healthHooks = append(b.healthHooks, handler)
}

// RunHealthCheck runs the registered health check hooks and returns an error if any health check fails.
//
// This function iterates through all registered health check hooks, executing each one within a specified timeout context.
// If any health check fails, it logs the failure and returns an error indicating which hook failed.
func (b *Base) RunHealthCheck(ctx context.Context) error {
	b.log.Debug(ctx).Msg("Starting health check")
	n := len(b.healthHooks)
	for i, hook := range b.healthHooks {
		name := reflect.TypeOf(hook).Elem().Name()
		b.log.Info(ctx).Msgf("Running health check %v of %v : %v", i+1, n, name)
		hookCtx, _ := context.WithTimeout(ctx, time.Second)
		result := make(chan error)
		go func() {
			result <- hook.HealthCheck(hookCtx)
		}()
		var err error
		select {
		case <-hookCtx.Done():
			err = context.DeadlineExceeded
		case err = <-result:
		}
		if err != nil {
			b.log.Error(ctx).Msg("health check failed for hook: " + name)
			return fmt.Errorf("Base.HealthCheck: %w", err)
		}
		b.log.Info(ctx).Msgf("Completed health check %v of %v : %v", i+1, n, name)
	}
	b.log.Debug(ctx).Msg("Completed health check")
	return nil
}
