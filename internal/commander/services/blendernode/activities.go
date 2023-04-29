package blendernode

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/flowshot-io/x/pkg/archiver"
	"go.beyondstorage.io/v5/types"
	"go.temporal.io/sdk/activity"
)

var (
	workingDir = "temp"
)

type Activities struct {
	storager types.Storager
	archiver *archiver.Archiver
}

func (a *Activities) DownloadArtifactActivity(ctx context.Context, artifact string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading artifact...", "Artifact", artifact)

	if !isArtifactFile(artifact) {
		return "", fmt.Errorf("downloadArtifactActivity only supports tar.gz files. %s is not a tar.gz file", artifact)
	}

	artifactPath, err := a.storagerRead(ctx, artifact, getArtifactLocalStorage(ctx))
	if err != nil {
		logger.Error("DownloadArtifactActivity failed to download artifact.", "Error", err)
		return "", err
	}

	logger.Info("DownloadArtifactActivity succeed.", "ArtifactPath", artifactPath)
	return artifactPath, nil
}

func (a *Activities) ExtractArtifactActivity(ctx context.Context, artifact string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Extracting artifact...", "Artifact", artifact)

	if !isArtifactFile(artifact) {
		return "", fmt.Errorf("extractArtifactActivity only supports tar.gz files. %s is not a tar.gz file", artifact)
	}

	dest := getWorkflowLocalStorage(ctx)
	err := a.archiver.Unarchive(artifact, dest)
	if err != nil {
		logger.Error("ExtractArtifactActivity failed to extract artifact.", "Error", err)
		return "", err
	}

	logger.Info("ExtractArtifactActivity succeed.", "ExtractedPath", dest)
	return dest, nil
}

func (a *Activities) RenderProjectActivity(ctx context.Context, projectDir string, startFrame int, endFrame int) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Rendering project...", "ProjectDir", projectDir, "StartFrame", startFrame, "EndFrame", endFrame)

	output, err := renderFile(ctx, projectDir, startFrame, endFrame)
	if err != nil {
		logger.Error("RenderFileActivity failed to render project.", "Error", err)
		return "", err
	}

	logger.Info("RenderFileActivity succeed.", "Output", output)
	return output, nil
}

func (a *Activities) CreateArtifactActivity(ctx context.Context, artifactName string, files []string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating artifact...", "ArtifactName", artifactName, "Files", files)

	if !validateArtifactName(artifactName) {
		logger.Error("CreateArtifactActivity failed to create artifact.", "Error", "Invalid artifact name.")
		return "", fmt.Errorf("invalid artifact name: %s", artifactName)
	}

	artifactOuput := filepath.Join(getArtifactLocalStorage(ctx), getArtifactFileName(artifactName))
	err := a.archiver.Archive(files, artifactOuput)
	if err != nil {
		logger.Error("CreateArtifactActivity failed to create artifact.", "Error", err)
		return "", err
	}

	logger.Info("CreateArtifactActivity succeed.", "Artifact", artifactOuput)
	return artifactOuput, nil
}

func (a *Activities) UploadArtifactActivity(ctx context.Context, artifact string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("uploadArtifactActivity begin.", "Artifact", artifact)

	if !isArtifactFile(artifact) {
		return fmt.Errorf("uploadArtifactActivity only supports tar.gz files. %s is not a tar.gz file", artifact)
	}

	err := a.storagerWrite(ctx, artifact, "artifacts")
	if err != nil {
		logger.Error("uploadArtifactActivity uploading failed.", "Error", err)
		return err
	}

	logger.Info("uploadArtifactActivity succeed.", "UploadedArtifact", artifact)
	return nil
}

func (a *Activities) CleanupActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("cleanupActivity begin.")

	workingDir := getWorkflowLocalStorage(ctx)
	err := os.RemoveAll(workingDir)
	if err != nil {
		logger.Error("cleanupActivity failed to remove working dir.", "Error", err)
		return err
	}

	logger.Info("cleanupActivity succeed.")
	return nil
}

func renderFile(ctx context.Context, workingDir string, frameStart int, frameEnd int) (string, error) {
	logger := activity.GetLogger(ctx)

	outputDir, err := os.MkdirTemp(workingDir, "*")
	if err != nil {
		return "", err
	}

	args := []string{
		"render",
		"-d", workingDir,
		"-s", fmt.Sprintf("%d", frameStart),
		"-e", fmt.Sprintf("%d", frameEnd),
		"-o", outputDir + "/{{.Project}}-#####",
	}

	cmd := exec.CommandContext(ctx, "flowshot", args...)
	logger.Info("renderFileActivity executing command.", "Command", cmd.String())

	err = cmd.Start()
	if err != nil {
		return "", err
	}
	defer cmd.Process.Kill()

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

OuterLoop:
	for {
		select {
		case <-ctx.Done():
			logger.Info("renderFileActivity command cancelled.")
			return "", ctx.Err()
		case err := <-done:
			if err != nil {
				return "", err
			}
			// Command has finished
			break OuterLoop
		default:
			activity.RecordHeartbeat(ctx, time.Now())
			time.Sleep(30 * time.Second)
		}
	}

	logger.Info("renderFileActivity command finished.")

	return outputDir, nil
}

func (a *Activities) storagerRead(ctx context.Context, filePath string, destPath string) (string, error) {
	path := filepath.Join(destPath, filepath.Base(filePath))

	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create directories: %w", err)
	}

	fmt.Println("Downloading file to: " + path)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = a.storager.ReadWithContext(ctx, filePath, file)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}

	return path, nil
}

func (a *Activities) storagerWrite(ctx context.Context, filePath string, destPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stat: %w", err)
	}

	_, err = a.storager.WriteWithContext(ctx, filepath.Join(destPath, filepath.Base(filePath)), file, stat.Size())
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func getWorkflowLocalStorage(ctx context.Context) string {
	return filepath.Join(workingDir, activity.GetInfo(ctx).WorkflowExecution.ID)
}

func getArtifactLocalStorage(ctx context.Context) string {
	return filepath.Join(getWorkflowLocalStorage(ctx), "artifacts")
}
