package blendernode

import (
	"fmt"
	"time"

	commanderactivities "github.com/flowshot-io/commander/internal/commander/temporalactivities"
	"github.com/flowshot-io/x/pkg/temporalactivities"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const Query = "blendernode-query"

type (
	BlenderNodeWorkflowInput struct {
		Artifact   string
		StartFrame int
		EndFrame   int
	}

	BlenderNodeWorkflowOutput struct {
		Result string
	}
)

func BlenderNodeWorkflow(ctx workflow.Context, request BlenderNodeWorkflowInput) (BlenderNodeWorkflowOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Minute,
		HeartbeatTimeout:    1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Render started", "Artifact", request.Artifact, "StartFrame", request.StartFrame, "EndFrame", request.EndFrame)

	outputArtifact, err := renderProjectArtifact(ctx, request.Artifact, request.StartFrame, request.EndFrame)
	if err != nil {
		logger.Error("Workflow failed.", "Error", err.Error())
		return BlenderNodeWorkflowOutput{Result: "ERR"}, err
	}

	logger.Info("Workflow completed.")

	return BlenderNodeWorkflowOutput{Result: outputArtifact}, nil
}

func renderProjectArtifact(ctx workflow.Context, projectArtifact string, startFrame int, endFrame int) (string, error) {
	so := &workflow.SessionOptions{
		CreationTimeout:  30 * time.Minute,
		ExecutionTimeout: 60 * time.Minute,
		HeartbeatTimeout: 1 * time.Minute,
	}
	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return "", err
	}
	defer workflow.CompleteSession(sessionCtx)

	var artifactAct *temporalactivities.ArtifactActivities
	var extractedDir string
	err = workflow.ExecuteActivity(sessionCtx, artifactAct.PullArtifact, projectArtifact, workflow.GetInfo(ctx).WorkflowExecution.ID).Get(sessionCtx, &extractedDir)
	if err != nil {
		return "", err
	}

	var blenderAct *commanderactivities.BlenderActivities
	var outputDir string
	err = workflow.ExecuteActivity(sessionCtx, blenderAct.RenderProjectActivity, extractedDir, startFrame, endFrame).Get(sessionCtx, &outputDir)
	if err != nil {
		return "", err
	}

	outputArtifactName := fmt.Sprintf("%s-%d-%d", workflow.GetInfo(ctx).WorkflowExecution.ID, startFrame, endFrame)
	var outputArtifact string
	err = workflow.ExecuteActivity(sessionCtx, artifactAct.PushArtifact, outputArtifactName, []string{outputDir}).Get(sessionCtx, &outputArtifact)
	if err != nil {
		return "", err
	}

	return outputArtifact, nil
}
