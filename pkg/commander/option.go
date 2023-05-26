package commander

import (
	"github.com/flowshot-io/commander/pkg/commander/config"
	"github.com/flowshot-io/x/pkg/logger"
	"github.com/flowshot-io/x/pkg/manager"
)

type (
	ServerOption interface {
		apply(*serverOptions)
	}

	applyFunc func(*serverOptions)
)

func (f applyFunc) apply(s *serverOptions) { f(s) }

// WithConfig sets a custom configuration
func WithConfig(cfg *config.Config) ServerOption {
	return applyFunc(func(s *serverOptions) {
		s.config = cfg
	})
}

// WithConfigLoader sets a custom configuration load
func WithConfigLoader(configDir string) ServerOption {
	return applyFunc(func(s *serverOptions) {
		s.configDir = configDir
	})
}

// WithServices indicates which supplied services (e.g. frontend, worker) within the server to start
func WithServices(names []string) ServerOption {
	return applyFunc(func(s *serverOptions) {
		s.serviceNames = make(map[manager.ServiceName]struct{})
		for _, name := range names {
			s.serviceNames[manager.ServiceName(name)] = struct{}{}
		}
	})
}

// InterruptOn interrupts server on the signal from server. If channel is nil Start() will block forever.
func InterruptOn(interruptCh <-chan interface{}) ServerOption {
	return applyFunc(func(s *serverOptions) {
		s.blockingStart = true
		s.interruptCh = interruptCh
	})
}

// WithLogger sets a custom logger
func WithLogger(logger logger.Logger) ServerOption {
	return applyFunc(func(s *serverOptions) {
		s.logger = logger
	})
}
