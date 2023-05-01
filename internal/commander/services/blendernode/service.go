package blendernode

import (
	"log"

	"github.com/flowshot-io/commander/internal/commander/temporalactivities"
	"github.com/flowshot-io/x/pkg/artifactservice"
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

func New(temporal client.Client, artifactClient artifactservice.ArtifactServiceClient, logger *log.Logger) *Service {
	worker := worker.New(temporal, Queue, worker.Options{
		EnableSessionWorker:               true,
		MaxConcurrentSessionExecutionSize: 1,
	})

	worker.RegisterWorkflow(BlenderNodeWorkflow)
	worker.RegisterActivity(temporalactivities.NewBlenderActivities())
	worker.RegisterActivity(temporalactivities.NewArtifactActivities(artifactClient))

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
