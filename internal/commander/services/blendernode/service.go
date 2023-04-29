package blendernode

import (
	"log"

	"github.com/flowshot-io/x/pkg/archiver"
	"go.beyondstorage.io/v5/types"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const Queue = "blendernode-queue"

type (
	Service struct {
		logger *log.Logger
		worker worker.Worker
	}
)

func New(temporal client.Client, storager types.Storager, logger *log.Logger) *Service {
	worker := worker.New(temporal, Queue, worker.Options{
		EnableSessionWorker:               true,
		MaxConcurrentSessionExecutionSize: 1,
	})

	worker.RegisterWorkflow(BlenderNodeWorkflow)
	worker.RegisterActivity(&Activities{
		storager: storager,
		archiver: &archiver.Archiver{},
	})

	return &Service{
		worker: worker,
		logger: logger,
	}
}

func (s *Service) Start() error {
	s.logger.Println("Starting blendernode service")

	err := s.worker.Start()
	if err != nil {
		s.logger.Println("Unable to start worker", map[string]interface{}{"Error": err.Error()})
		return err
	}

	return nil
}

func (s *Service) Stop() error {
	s.logger.Println("Stopping blendernode service")
	s.worker.Stop()
	s.logger.Println("blendernode service stopped")

	return nil
}
