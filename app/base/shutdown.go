package base

import (
	"context"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

// AwaitShutdownCompletion waits for graceful shutdown.
//
// This function blocks until the shutdown process, including all registered shutdown hooks, is complete.
func (b *Base) AwaitShutdownCompletion() {
	b.shutdownWg.Wait()
}

// StartSignalMonitor starts monitoring for OS signals such as SIGTERM and SIGINT and initiates shutdown on receiving them.
//
// This function sets up a channel to receive OS signals and starts a goroutine to monitor those signals.
// When a signal is received, it triggers the server shutdown process.
func (b *Base) StartSignalMonitor(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	go b.monitorSignals(ctx, c)
	return nil
}

// RegisterOnShutdownHook registers a shutdown hook to be executed during server shutdown.
//
// This function appends the provided shutdown handler to the list of shutdown hooks in the BaseApp.
func (b *Base) RegisterOnShutdownHook(handler ShutdownHook) {
	b.shutdownHooks = append(b.shutdownHooks, handler)
}

// Shutdown gracefully shuts down the server by executing registered shutdown hooks.
//
// This function iterates through all registered shutdown hooks, executing each one within a specified timeout context.
// It logs the progress of the shutdown process and ensures all hooks are processed before completing the shutdown.
func (b *Base) Shutdown(ctx context.Context) {
	b.log.Info(ctx).Msg("Gracefully shutting down")
	hooksCount := len(b.shutdownHooks)
	for i, hook := range b.shutdownHooks {
		name := reflect.TypeOf(hook).Elem().Name()
		shutdownCtx, _ := context.WithTimeout(ctx, time.Second*2)
		b.log.Info(ctx).Msgf("closing hook %v of %v - %v", i+1, hooksCount, name)
		b.processShutdownHook(shutdownCtx, name, hook)
		b.log.Info(ctx).Msgf("closed hook %v of %v - %v", i+1, hooksCount, name)
	}
	b.log.Info(ctx).Msg("shutdown completed")
	b.shutdownWg.Done()
}

// processShutdownHook executes the shutdown logic for a single shutdown hook.
//
// This function runs the shutdown logic for the provided handler within a deferred recovery block to handle any panics.
// It logs any errors that occur during the shutdown process of the handler.
func (b *Base) processShutdownHook(ctx context.Context, name string, handler ShutdownHook) {
	defer func() {
		if rec := recover(); rec != nil {
			b.log.Error(ctx).Any("panic", rec).Msg("panic closing: " + name)
		}
	}()
	err := handler.Close(ctx)
	if err != nil {
		b.log.Error(ctx).Err(err).Msg("error closing: " + name)
		return
	}
}

// monitorSignals monitors OS signals and initiates server shutdown upon receiving them.
//
// This function blocks until an OS signal is received on the provided channel, then calls the Shutdown method.
func (b *Base) monitorSignals(ctx context.Context, ch chan os.Signal) {
	<-ch
	b.Shutdown(ctx)
}
