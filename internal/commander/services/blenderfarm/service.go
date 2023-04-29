package blenderfarm

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const Queue = "blenderfarm-queue"

type (
	Service struct {
		logger *log.Logger
		worker worker.Worker
	}
)

func New(temporal client.Client, logger *log.Logger) *Service {
	worker := worker.New(temporal, Queue, worker.Options{})

	worker.RegisterWorkflow(BlenderFarmWorkflow)

	return &Service{
		worker: worker,
		logger: logger,
	}
}

func (s *Service) Start() error {
	s.logger.Println("Starting blenderfarm service")

	err := s.worker.Start()
	if err != nil {
		s.logger.Println("Unable to start worker", map[string]interface{}{"Error": err.Error()})
		return err
	}

	return nil
}

func (s *Service) Stop() error {
	s.logger.Println("Stopping blenderfarm service")
	s.worker.Stop()
	s.logger.Println("blenderfarm service stopped")

	return nil
}
