package factory

import (
	"fmt"
	"time"

	_ "github.com/flowshot-io/polystore/pkg/services/fs"
	_ "github.com/flowshot-io/polystore/pkg/services/s3"

	"github.com/flowshot-io/polystore/pkg/types"
	"go.temporal.io/sdk/client"

	"github.com/flowshot-io/polystore/pkg/services"
	"github.com/flowshot-io/x/pkg/artifactservice"
	"github.com/flowshot-io/x/pkg/logger"
	"github.com/flowshot-io/x/pkg/temporallogger"
)

type Factory interface {
	TemporalClient(hostPort string) (client.Client, error)
	ArtifactClient(connectionString string) (artifactservice.ArtifactServiceClient, error)
}

type RetryFactory struct {
	maxRetries    int
	retryInterval time.Duration
	logger        logger.Logger
}

type OptionsFunc func(*RetryFactory)

func WithMaxRetries(maxRetries int) OptionsFunc {
	return func(f *RetryFactory) {
		f.maxRetries = maxRetries
	}
}

func WithRetryInterval(retryInterval time.Duration) OptionsFunc {
	return func(f *RetryFactory) {
		f.retryInterval = retryInterval
	}
}

func WithLogger(logger logger.Logger) OptionsFunc {
	return func(f *RetryFactory) {
		f.logger = logger
	}
}

func New(opts ...OptionsFunc) Factory {
	factory := &RetryFactory{
		maxRetries:    5,
		retryInterval: 5 * time.Second,
		logger:        logger.NoOp(),
	}

	for _, opt := range opts {
		opt(factory)
	}

	return factory
}

func (f *RetryFactory) ArtifactClient(connectionString string) (artifactservice.ArtifactServiceClient, error) {
	var err error
	var store types.Storage
	for i := 0; i < f.maxRetries; i++ {
		store, err = services.NewStorageFromString(connectionString)
		if err != nil {
			f.logger.Error(fmt.Sprintf("Attempt %d: unable to create Artifact client: %v", i+1, err))
			time.Sleep(f.retryInterval * time.Duration(i+1))
			continue
		}
		return artifactservice.New(artifactservice.Options{
			Store: store,
		})
	}
	return nil, fmt.Errorf("unable to create Artifact client after %d attempts", f.maxRetries)
}

func (f *RetryFactory) TemporalClient(hostPort string) (client.Client, error) {
	var err error
	var temporal client.Client
	for i := 0; i < f.maxRetries; i++ {
		temporal, err = client.Dial(client.Options{
			HostPort: hostPort,
			Logger:   temporallogger.New(f.logger),
		})
		if err != nil {
			f.logger.Error(fmt.Sprintf("Attempt %d: unable to create Temporal client: %v", i+1, err))
			time.Sleep(f.retryInterval * time.Duration(i+1))
			continue
		}
		return temporal, nil
	}

	return nil, fmt.Errorf("unable to create Temporal client after %d attempts", f.maxRetries)
}
