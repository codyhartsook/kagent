package client

import (
	"context"

	apiv1alpha1 "github.com/kagent-dev/kagent/go/api/gen/kagent/api/v1alpha1"
)

// AgentInstance defines AgentInstance lifecycle operations.
type AgentInstance interface {
	CreateAgentInstance(context.Context, *apiv1alpha1.CreateAgentInstanceRequest) (*apiv1alpha1.AgentInstance, error)
	GetAgentInstance(context.Context, *apiv1alpha1.GetAgentInstanceRequest) (*apiv1alpha1.AgentInstance, error)
	ListAgentInstances(context.Context, *apiv1alpha1.ListAgentInstancesRequest) (*apiv1alpha1.ListAgentInstancesResponse, error)
}

type agentInstanceClient struct {
	client *BaseClient
}

// NewAgentInstanceClient creates an AgentInstance client.
func NewAgentInstanceClient(client *BaseClient) AgentInstance {
	return &agentInstanceClient{client: client}
}

func (c *agentInstanceClient) CreateAgentInstance(ctx context.Context, request *apiv1alpha1.CreateAgentInstanceRequest) (*apiv1alpha1.AgentInstance, error) {
	client, err := c.serviceClient()
	if err != nil {
		return nil, err
	}
	callContext, cancel := c.client.grpcCallContext(ctx)
	defer cancel()
	response, err := client.CreateAgentInstance(callContext, request)
	if err != nil {
		return nil, err
	}
	return response.GetAgentInstance(), nil
}

func (c *agentInstanceClient) GetAgentInstance(ctx context.Context, request *apiv1alpha1.GetAgentInstanceRequest) (*apiv1alpha1.AgentInstance, error) {
	client, err := c.serviceClient()
	if err != nil {
		return nil, err
	}
	callContext, cancel := c.client.grpcCallContext(ctx)
	defer cancel()
	response, err := client.GetAgentInstance(callContext, request)
	if err != nil {
		return nil, err
	}
	return response.GetAgentInstance(), nil
}

func (c *agentInstanceClient) ListAgentInstances(ctx context.Context, request *apiv1alpha1.ListAgentInstancesRequest) (*apiv1alpha1.ListAgentInstancesResponse, error) {
	client, err := c.serviceClient()
	if err != nil {
		return nil, err
	}
	callContext, cancel := c.client.grpcCallContext(ctx)
	defer cancel()
	return client.ListAgentInstances(callContext, request)
}

func (c *agentInstanceClient) serviceClient() (apiv1alpha1.AgentInstanceServiceClient, error) {
	connection, err := c.client.grpcConnection()
	if err != nil {
		return nil, err
	}
	return apiv1alpha1.NewAgentInstanceServiceClient(connection), nil
}
