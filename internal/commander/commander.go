package commander

import (
	"fmt"

	_ "github.com/flowshot-io/polystore/pkg/services/fs"
	_ "github.com/flowshot-io/polystore/pkg/services/s3"

	"github.com/flowshot-io/commander/internal/commander/primitives"
	"github.com/flowshot-io/commander/internal/commander/services/blenderfarm"
	"github.com/flowshot-io/commander/internal/commander/services/blendernode"
	"github.com/flowshot-io/commander/internal/commander/services/frontend"
	"github.com/flowshot-io/polystore/pkg/services"
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
		ServerOptions *serverOptions
		services      *manager.ServiceManager
	}
)

func New(opts ...ServerOption) (*Commander, error) {
	so, err := ServerOptions(opts)
	if err != nil {
		return nil, err
	}

	topts := client.Options{
		HostPort: so.config.Global.Temporal.Host,
	}

	temporal, err := client.Dial(topts)
	if err != nil {
		return nil, fmt.Errorf("unable to create Temporal client: %w", err)
	}

	store, err := services.NewStorageFromString(so.config.Global.Storage.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to create Storage client: %w", err)
	}

	artifact, err := artifactservice.New(artifactservice.Options{
		Store: store,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Artifact client: %w", err)
	}

	sm := manager.New()

	for serviceName := range so.serviceNames {
		switch serviceName {
		case primitives.FrontendService:
			sm.Add(serviceName, frontend.New(temporal, so.logger))
		case primitives.BlenderFarmService:
			sm.Add(serviceName, blenderfarm.New(temporal, so.logger))
		case primitives.BlenderNodeService:
			sm.Add(serviceName, blendernode.New(temporal, artifact, so.logger))
		}
	}

	return &Commander{
		ServerOptions: so,
		services:      sm,
	}, nil
}

func (c *Commander) Start() {
	c.services.Start()
}

func (c *Commander) Stop() {
	c.services.Stop()
}

func ServerOptions(opts []ServerOption) (*serverOptions, error) {
	so := newServerOptions(opts)

	err := so.loadAndValidate()
	if err != nil {
		return nil, err
	}

	return so, nil
}
