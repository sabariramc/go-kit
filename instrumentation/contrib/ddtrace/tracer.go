// Package ddtrace is the implementation of instrumentation.Tracer for Datadog.
package ddtrace

import (
	"sync"

	ddtrace "github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
)

// tracer is an empty struct implementing the instrumentation.Tracer interface for Datadog.
type Tracer struct {
}

// Init initializes the Datadog tracer and returns an instance of instrumentation.Tracer.
// This function starts the Datadog tracer.
func Init() (*Tracer, error) {
	sy := &sync.Once{}
	sy.Do(func() {
		ddtrace.Start()
	})
	return &Tracer{}, nil
}

// ShutDown stops the Datadog tracer. This function should be called to properly
// shut down the tracer and flush any remaining traces.
func ShutDown() {
	sy := &sync.Once{}
	sy.Do(func() {
		ddtrace.Stop()
	})
}
