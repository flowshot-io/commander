package temporalactivities

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"go.temporal.io/sdk/activity"
)

type BlenderActivities struct {
}

func NewBlenderActivities() *BlenderActivities {
	return &BlenderActivities{}
}

func (a *BlenderActivities) RenderProjectActivity(ctx context.Context, workingDir string, startFrame int, endFrame int) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Rendering project...", "WorkingDir", workingDir, "StartFrame", startFrame, "EndFrame", endFrame)

	output, err := renderFile(ctx, workingDir, startFrame, endFrame)
	if err != nil {
		logger.Error("RenderFileActivity failed to render project.", "Error", err)
		return "", err
	}

	logger.Info("RenderFileActivity succeed.", "Output", output)
	return output, nil
}

func renderFile(ctx context.Context, workingDir string, frameStart int, frameEnd int) (string, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("renderFileActivity starting...", "WorkingDir", workingDir, "FrameStart", frameStart, "FrameEnd", frameEnd)

	args := []string{
		"render",
		"-d", workingDir,
		"-s", fmt.Sprintf("%d", frameStart),
		"-e", fmt.Sprintf("%d", frameEnd),
	}

	cmd := exec.CommandContext(ctx, "rocketblend", args...)
	logger.Info("renderFileActivity executing command.", "Command", cmd.String())

	err := cmd.Start()
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

	return filepath.Join(workingDir, "output"), nil
}
