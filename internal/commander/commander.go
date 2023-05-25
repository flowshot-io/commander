package commander

import (
	"fmt"

	_ "github.com/flowshot-io/polystore/pkg/services/fs"
	_ "github.com/flowshot-io/polystore/pkg/services/s3"

	"github.com/flowshot-io/commander/internal/commander/factory"
	"github.com/flowshot-io/commander/internal/commander/primitives"
	"github.com/flowshot-io/commander/internal/commander/services/blenderfarm"
	"github.com/flowshot-io/commander/internal/commander/services/blendernode"
	"github.com/flowshot-io/commander/internal/commander/services/frontend"
	"github.com/flowshot-io/x/pkg/artifactservice"
	"github.com/flowshot-io/x/pkg/manager"
	"go.temporal.io/sdk/client"
)

var (
	Services = []string{
		string(primitives.FrontendService),
		string(primitives.BlenderNodeService),
		string(primitives.BlenderFarmService),
	}
)

type (
	Commander struct {
		serverOptions *serverOptions
		factory       factory.Factory
		services      manager.ServiceController
	}
)

func New(opts ...ServerOption) (manager.Service, error) {
	so, err := ServerOptions(opts)
	if err != nil {
		return nil, err
	}

	return &Commander{
		serverOptions: so,
		services:      manager.New(&manager.Options{Logger: so.logger}),
		factory:       factory.New(factory.WithLogger(so.logger)),
	}, nil
}

func (c *Commander) Start() error {
	temporalClient, err := c.factory.TemporalClient(c.serverOptions.config.Global.Temporal.Host)
	if err != nil {
		return fmt.Errorf("unable to create temporal client: %w", err)
	}

	artifactClient, err := c.factory.ArtifactClient(c.serverOptions.config.Global.Storage.ConnectionString)
	if err != nil {
		return fmt.Errorf("unable to create artifact client: %w", err)
	}

	err = c.initServices(temporalClient, artifactClient)
	if err != nil {
		return fmt.Errorf("unable to init services: %w", err)
	}

	return c.services.Start()
}

func (c *Commander) Stop() error {
	return c.services.Stop()
}

func ServerOptions(opts []ServerOption) (*serverOptions, error) {
	so := newServerOptions(opts)

	err := so.loadAndValidate()
	if err != nil {
		return nil, err
	}

	return so, nil
}

func (c *Commander) initServices(temporalClient client.Client, artifactClient artifactservice.ArtifactServiceClient) error {
	for serviceName := range c.serverOptions.serviceNames {
		switch serviceName {
		case primitives.FrontendService:
			srv, err := frontend.New(frontend.Options{TemporalClient: temporalClient, Logger: c.serverOptions.logger})
			if err != nil {
				return fmt.Errorf("unable to create frontend service: %w", err)
			}
			err = c.services.Add(serviceName, srv)
			if err != nil {
				return fmt.Errorf("unable to add frontend service: %w", err)
			}
		case primitives.BlenderFarmService:
			srv, err := blenderfarm.New(blenderfarm.Options{TemporalClient: temporalClient, Logger: c.serverOptions.logger})
			if err != nil {
				return fmt.Errorf("unable to create blenderfarm service: %w", err)
			}
			err = c.services.Add(serviceName, srv)
			if err != nil {
				return fmt.Errorf("unable to add blenderfarm service: %w", err)
			}
		case primitives.BlenderNodeService:
			srv, err := blendernode.New(blendernode.Options{TemporalClient: temporalClient, ArtifactClient: artifactClient, Logger: c.serverOptions.logger})
			if err != nil {
				return fmt.Errorf("unable to create blendernode service: %w", err)
			}

			err = c.services.Add(serviceName, srv)
			if err != nil {
				return fmt.Errorf("unable to add blendernode service: %w", err)
			}
		}
	}

	return nil
}
