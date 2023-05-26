package blenderfarm

import (
	"fmt"

	"github.com/flowshot-io/x/pkg/logger"
	"github.com/flowshot-io/x/pkg/manager"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const Queue = "blenderfarm-queue"

type (
	Options struct {
		TemporalClient client.Client
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

	worker := worker.New(opts.TemporalClient, Queue, worker.Options{})

	worker.RegisterWorkflow(BlenderFarmWorkflow)

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
