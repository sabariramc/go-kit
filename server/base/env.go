package base

import (
	"sync"

	"github.com/sabariramc/go-kit/env"
)

const (
	EnvServiceName = "SERVICE_NAME"
)

var GetServiceName = sync.OnceValue(
	func() string {
		return env.Get(EnvServiceName, "default-service")
	},
)
