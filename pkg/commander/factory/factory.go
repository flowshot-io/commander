package factory

import (
	"context"
	"fmt"
	"math"
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
	TemporalClient(ctx context.Context, hostPort string) (client.Client, error)
	ArtifactClient(ctx context.Context, connectionString string) (artifactservice.ArtifactServiceClient, error)
}

type RetryFactory struct {
	maxRetries    int
	retryInterval time.Duration
	maxTotalTime  time.Duration
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

func WithMaxTotalTime(maxTotalTime time.Duration) OptionsFunc {
	return func(f *RetryFactory) {
		f.maxTotalTime = maxTotalTime
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
		maxTotalTime:  1 * time.Minute,
		logger:        logger.NoOp(),
	}

	for _, opt := range opts {
		opt(factory)
	}

	return factory
}

func (f *RetryFactory) ArtifactClient(ctx context.Context, connectionString string) (artifactservice.ArtifactServiceClient, error) {
	var store types.Storage
	err := f.retry(ctx, func() error {
		var err error
		store, err = services.NewStorageFromString(connectionString)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create Artifact client: %v", err)
	}

	return artifactservice.New(artifactservice.Options{
		Store: store,
	})
}

func (f *RetryFactory) TemporalClient(ctx context.Context, hostPort string) (client.Client, error) {
	var temporal client.Client
	err := f.retry(ctx, func() error {
		var err error
		temporal, err = client.Dial(client.Options{
			HostPort: hostPort,
			Logger:   temporallogger.New(f.logger),
		})
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("unable to create Temporal client: %v", err)
	}

	return temporal, nil
}

func (f *RetryFactory) retry(ctx context.Context, call func() error) error {
	elapsedTime := time.Duration(0)
	for i := 0; i < f.maxRetries && elapsedTime < f.maxTotalTime; i++ {
		err := call()
		if err == nil {
			return nil
		}

		sleepDuration := f.retryInterval * time.Duration(math.Pow(2, float64(i)))
		elapsedTime += sleepDuration
		if elapsedTime > f.maxTotalTime {
			sleepDuration -= elapsedTime - f.maxTotalTime
		}

		f.logger.Error(fmt.Sprintf("Attempt %d: error: %v", i+1, err))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepDuration):
			// Continue retrying.
		}
	}

	return fmt.Errorf("retrying failed after %d attempts and %s elapsed time", f.maxRetries, elapsedTime)
}
