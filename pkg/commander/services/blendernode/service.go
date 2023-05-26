package blendernode

import (
	"fmt"

	commanderactivities "github.com/flowshot-io/commander/pkg/commander/temporalactivities"
	"github.com/flowshot-io/x/pkg/artifactservice"
	"github.com/flowshot-io/x/pkg/logger"
	"github.com/flowshot-io/x/pkg/manager"
	"github.com/flowshot-io/x/pkg/temporalactivities"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const Queue = "blendernode-queue"

type (
	Options struct {
		TemporalClient client.Client
		ArtifactClient artifactservice.ArtifactServiceClient
		Logger         logger.Logger
	}

	Service struct {
		logger logger.Logger
		worker worker.Worker
	}
)

func New(opts Options) (manager.Service, error) {
	if opts.Logger == nil {
		opts.Logger = logger.NoOp()
	}

	if opts.TemporalClient == nil {
		return nil, fmt.Errorf("temporal client is required")
	}

	if opts.ArtifactClient == nil {
		return nil, fmt.Errorf("artifact client is required")
	}

	worker := worker.New(opts.TemporalClient, Queue, worker.Options{
		EnableSessionWorker:               true,
		MaxConcurrentSessionExecutionSize: 1,
	})

	worker.RegisterWorkflow(BlenderNodeWorkflow)
	worker.RegisterActivity(commanderactivities.NewBlenderActivities())
	worker.RegisterActivity(temporalactivities.NewArtifactActivities(opts.ArtifactClient))

	return &Service{
		worker: worker,
		logger: opts.Logger,
	}, nil
}

func (s *Service) Start() error {
	err := s.worker.Start()
	if err != nil {
		s.logger.Error("Unable to start worker", map[string]interface{}{"Error": err.Error()})
		return err
	}

	return nil
}

func (s *Service) Stop() error {
	s.worker.Stop()
	return nil
}
