package base

import "context"

// ShutdownHook defines an interface for the graceful shutdown of different resources used by the app.
//
// This interface extends the Name interface and requires the following method:
//   - Shutdown(ctx context.Context) error
type ShutdownHook interface {
	Close(ctx context.Context) error
}

// HealthCheckHook defines an interface for health checks of different resources used by the app.
//
// This interface extends the Name interface and requires the following method:
//   - HealthCheck(ctx context.Context) error
type HealthCheckHook interface {
	HealthCheck(ctx context.Context) error
}

// RegisterHooks registers the provided hooks to the BaseApp.
//
// This function checks the type of the provided hook and registers it as a HealthCheckHook and/or ShutdownHook if it implements the respective interface.
func (b *Base) RegisterHooks(hook any) {
	if hHook, ok := hook.(HealthCheckHook); ok {
		b.RegisterHealthCheckHook(hHook)
	}
	if sHook, ok := hook.(ShutdownHook); ok {
		b.RegisterOnShutdownHook(sHook)
	}
}
