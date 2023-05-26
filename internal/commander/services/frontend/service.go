package frontend

import (
	"fmt"
	"net"

	"github.com/flowshot-io/commander-client-go/commanderservice/v1"
	"github.com/flowshot-io/x/pkg/logger"
	"github.com/flowshot-io/x/pkg/manager"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

type Options struct {
	TemporalClient client.Client
	Logger         logger.Logger
}

type Service struct {
	Server  *grpc.Server
	Logger  logger.Logger
	Port    int
	ErrChan chan error
}

type server struct {
	commanderservice.CommanderServiceServer
	temporal client.Client
}

func New(opts Options) (manager.Service, error) {
	if opts.Logger == nil {
		opts.Logger = logger.NoOp()
	}

	if opts.TemporalClient == nil {
		return nil, fmt.Errorf("temporal client is required")
	}

	srv := grpc.NewServer()
	commanderservice.RegisterCommanderServiceServer(srv, &server{
		temporal: opts.TemporalClient,
	})

	s := &Service{
		Server:  srv,
		Logger:  opts.Logger,
		Port:    50051,
		ErrChan: make(chan error),
	}

	return s, nil
}

func (s *Service) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	go func() {
		if err := s.Server.Serve(lis); err != nil {
			s.ErrChan <- fmt.Errorf("failed to serve: %v", err)
		}
	}()

	return nil
}

func (s *Service) Stop() error {
	s.Server.GracefulStop()
	close(s.ErrChan)
	return nil
}
