package frontend

import (
	"context"

	"github.com/flowshot-io/commander-client-go/commanderservice/v1"
	"github.com/flowshot-io/commander/pkg/commander/services/blenderfarm"
	"go.temporal.io/sdk/client"
)

func (s *server) GetBlenderFarmWorkflow(ctx context.Context, req *commanderservice.GetBlenderFarmWorkflowRequest) (*commanderservice.GetBlenderFarmWorkflowResponse, error) {
	response, err := s.temporal.QueryWorkflow(ctx, req.Id, "", blenderfarm.Query)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	if err := response.Get(&res); err != nil {
		return nil, err
	}

	return &commanderservice.GetBlenderFarmWorkflowResponse{
		Status: &commanderservice.RenderStatus{
			Id: req.Id,
		},
	}, nil
}

func (s *server) ListBlenderFarmWorkflows(ctx context.Context, req *commanderservice.ListBlenderFarmWorkflowsRequest) (*commanderservice.ListBlenderFarmWorkflowsResponse, error) {
	return &commanderservice.ListBlenderFarmWorkflowsResponse{}, nil
}

func (s *server) CreateBlenderFarmWorkflow(ctx context.Context, req *commanderservice.CreateBlenderFarmWorkflowRequest) (*commanderservice.CreateBlenderFarmWorkflowResponse, error) {
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: blenderfarm.Queue,
	}

	we, err := s.temporal.ExecuteWorkflow(ctx, workflowOptions, blenderfarm.BlenderFarmWorkflow, blenderfarm.BlenderFarmWorkflowInput{
		Artifact:   req.File,
		StartFrame: int(req.StartFrame),
		EndFrame:   int(req.EndFrame),
		BatchSize:  int(req.BatchSize),
	})
	if err != nil {
		return nil, err
	}

	return &commanderservice.CreateBlenderFarmWorkflowResponse{
		Id: we.GetID(),
	}, nil
}

func (s *server) CancelBlenderFarmWorkflow(ctx context.Context, req *commanderservice.CancelBlenderFarmWorkflowRequest) (*commanderservice.CancelBlenderFarmWorkflowResponse, error) {
	err := s.temporal.CancelWorkflow(ctx, req.Id, "")
	if err != nil {
		return nil, err
	}

	return &commanderservice.CancelBlenderFarmWorkflowResponse{}, nil
}
