package commander

import (
	"fmt"

	"github.com/flowshot-io/commander/pkg/commander/config"
	"github.com/flowshot-io/x/pkg/logger"
	"github.com/flowshot-io/x/pkg/manager"
	"golang.org/x/exp/slices"
)

type (
	serverOptions struct {
		serviceNames map[manager.ServiceName]struct{}

		config    *config.Config
		configDir string

		interruptCh   <-chan interface{}
		blockingStart bool

		logger logger.Logger
	}
)

func newServerOptions(opts []ServerOption) *serverOptions {
	so := &serverOptions{
		// Set defaults here.
	}
	for _, opt := range opts {
		opt.apply(so)
	}

	return so
}

func (so *serverOptions) loadAndValidate() error {
	for serviceName := range so.serviceNames {
		if !slices.Contains(Services, string(serviceName)) {
			return fmt.Errorf("invalid service %q in service list %v", serviceName, so.serviceNames)
		}
	}

	if so.config == nil {
		conf, err := config.LoadConfig(so.configDir)
		if err != nil {
			return fmt.Errorf("unable to load config: %w", err)
		}

		so.config = conf
	}

	err := so.validateConfig()
	if err != nil {
		return fmt.Errorf("config validation error: %w", err)
	}

	return nil
}

func (so *serverOptions) validateConfig() error {
	if err := so.config.Validate(); err != nil {
		return err
	}

	return nil
}
