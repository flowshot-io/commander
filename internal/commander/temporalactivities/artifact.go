package temporalactivities

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flowshot-io/x/pkg/artifact"
	"github.com/flowshot-io/x/pkg/artifactservice"
	"go.temporal.io/sdk/activity"
)

var (
	workingDir = "artifacts"
)

type ArtifactActivities struct {
	artifactClient artifactservice.ArtifactServiceClient
}

func NewArtifactActivities(artifactClient artifactservice.ArtifactServiceClient) *ArtifactActivities {
	return &ArtifactActivities{
		artifactClient: artifactClient,
	}
}

func (a *ArtifactActivities) PullArtifactActivity(ctx context.Context, artifactName string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading artifact...", "Artifact", artifactName)

	artifact, err := a.artifactClient.GetArtifact(ctx, artifactName)
	if err != nil {
		logger.Error("PullArtifactActivity failed to get artifact.", "Error", err)
		return "", err
	}

	artifactPath := getArtifactExtractionPath(ctx)
	artifact.ExtractToDirectory(artifactPath)

	logger.Info("PullArtifactActivity succeed.", "ExtractedTo", artifactPath)
	return artifactPath, nil
}

func (a *ArtifactActivities) PushArtifactActivity(ctx context.Context, artifactName string, files []string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating artifact...", "ArtifactName", artifactName, "Files", files)

	artifact, err := artifact.NewWithPaths(artifactName, files)
	if err != nil {
		logger.Error("PushArtifactActivity failed to create artifact.", "Error", err)
		return "", err
	}

	err = a.artifactClient.UploadArtifact(ctx, artifact)
	if err != nil {
		logger.Error("PushArtifactActivity failed to upload artifact.", "Error", err)
		return "", err
	}

	logger.Info("PushArtifactActivity succeed.", "PushedArtifact", artifact.GetName())
	return artifact.GetName(), nil
}

func (a *ArtifactActivities) CleanupArtifactActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("CleanupArtifactActivity begin.")

	err := os.RemoveAll(getArtifactExtractionPath(ctx))
	if err != nil {
		logger.Error("CleanupArtifactActivity failed to remove working dir.", "Error", err)
		return err
	}

	logger.Info("CleanupArtifactActivity succeed.")
	return nil
}

func getArtifactExtractionPath(ctx context.Context) string {
	return filepath.Join(workingDir, activity.GetInfo(ctx).WorkflowExecution.RunID)
}
