package blenderfarm

import (
	"time"

	"github.com/flowshot-io/commander/internal/commander/services/blendernode"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const Query = "blenderfarm-query"

type (
	BlenderFarmWorkflowInput struct {
		Artifact   string
		StartFrame int
		EndFrame   int
		BatchSize  int
	}

	BlenderFarmWorkflowOutput struct {
		Result []string
	}
)

func BlenderFarmWorkflow(ctx workflow.Context, request BlenderFarmWorkflowInput) (BlenderFarmWorkflowOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Minute,
		HeartbeatTimeout:    1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
		},
	}

	cwo := workflow.ChildWorkflowOptions{
		TaskQueue: blendernode.Queue,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: 10 * time.Second,
		},
	}

	ctx = workflow.WithChildOptions(ctx, cwo)
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Render started", "Artifact", request.Artifact, "StartFrame", request.StartFrame, "EndFrame", request.EndFrame, "BatchSize", request.BatchSize)

	if (request.BatchSize <= 0) || (request.BatchSize > (request.EndFrame - request.StartFrame + 1)) {
		request.BatchSize = request.EndFrame - request.StartFrame + 1
	}

	// Calculate the number of batches required based on the batch size and frame range.
	numBatches := (request.EndFrame - request.StartFrame + 1) / request.BatchSize
	if (request.EndFrame-request.StartFrame+1)%request.BatchSize != 0 {
		numBatches++
	}

	// Create a channel to receive the output of child workflows.
	resultCh := workflow.NewChannel(ctx)

	// Start child workflows for each batch.
	for i := 0; i < numBatches; i++ {
		startFrame := request.StartFrame + i*request.BatchSize
		endFrame := request.StartFrame + (i+1)*request.BatchSize - 1
		if endFrame > request.EndFrame {
			endFrame = request.EndFrame
		}

		childWorkflow := workflow.ExecuteChildWorkflow(ctx, blendernode.BlenderNodeWorkflow, blendernode.BlenderNodeWorkflowInput{
			Artifact:   request.Artifact,
			StartFrame: startFrame,
			EndFrame:   endFrame,
		})

		// Send the workflow execution result to the channel.
		workflow.Go(ctx, func(ctx workflow.Context) {
			var childWorkflowOutput blendernode.BlenderNodeWorkflowOutput
			err := childWorkflow.Get(ctx, &childWorkflowOutput)
			if err != nil {
				resultCh.Send(ctx, blendernode.BlenderNodeWorkflowOutput{Result: "ERR"})
				return
			}
			resultCh.Send(ctx, childWorkflowOutput)
		})
	}

	// Wait for all child workflows to complete.
	var results []blendernode.BlenderNodeWorkflowOutput
	for i := 0; i < numBatches; i++ {
		var result blendernode.BlenderNodeWorkflowOutput
		resultCh.Receive(ctx, &result)
		results = append(results, result)
	}

	// Combine the output of all child workflows.
	var output []string
	for _, result := range results {
		// if result.Result == "ERR" {
		// 	return BlenderFarmWorkflowOutput{Message: "ERR"}, errors.New("one or more child workflows failed")
		// }
		output = append(output, result.Result)
	}

	return BlenderFarmWorkflowOutput{Result: output}, nil
}
