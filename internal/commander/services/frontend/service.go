package frontend

import (
	"fmt"
	"log"
	"net"

	"github.com/flowshot-io/commander-client-go/commanderservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

type Service struct {
	Server *grpc.Server
	Logger *log.Logger
	Port   int
}

type server struct {
	commanderservice.CommanderServiceServer
	temporal client.Client
}

func New(temporal client.Client, logger *log.Logger) *Service {
	srv := grpc.NewServer()
	commanderservice.RegisterCommanderServiceServer(srv, &server{
		temporal: temporal,
	})

	s := &Service{
		Server: srv,
		Logger: logger,
		Port:   50051,
	}

	return s
}

func (s *Service) Start() error {
	s.Logger.Printf("Starting frontend service on port %d", s.Port)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	if err := s.Server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (s *Service) Stop() error {
	s.Logger.Println("Stopping frontend service")

	s.Server.GracefulStop()

	s.Logger.Println("Frontend service stopped")
	return nil
}
