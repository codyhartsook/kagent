package client

import (
	"context"
	"iter"

	a2atype "github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	"github.com/a2aproject/a2a-go/v2/a2agrpc/v1"
	"github.com/a2aproject/a2a-go/v2/a2apb/v1"
	apia2a "github.com/kagent-dev/kagent/go/api/a2a"
)

// A2A defines public AgentInstance interaction operations.
type A2A interface {
	SendMessage(context.Context, string, string, *a2atype.SendMessageRequest) (a2atype.SendMessageResult, error)
	SendStreamingMessage(context.Context, string, string, *a2atype.SendMessageRequest) iter.Seq2[a2atype.Event, error]
}

type a2aClient struct {
	client *BaseClient
}

// NewA2AClient creates a public A2A client.
func NewA2AClient(client *BaseClient) A2A {
	return &a2aClient{client: client}
}

func (c *a2aClient) SendMessage(ctx context.Context, namespace, instanceID string, request *a2atype.SendMessageRequest) (a2atype.SendMessageResult, error) {
	transport, err := c.transport()
	if err != nil {
		return nil, err
	}
	callContext, cancel := c.client.grpcCallContext(ctx)
	defer cancel()
	return transport.SendMessage(callContext, routeParams(c.client.UserID, namespace, instanceID), request)
}

func (c *a2aClient) SendStreamingMessage(ctx context.Context, namespace, instanceID string, request *a2atype.SendMessageRequest) iter.Seq2[a2atype.Event, error] {
	transport, err := c.transport()
	if err != nil {
		return func(yield func(a2atype.Event, error) bool) { yield(nil, err) }
	}
	return func(yield func(a2atype.Event, error) bool) {
		transport.SendStreamingMessage(ctx, routeParams(c.client.UserID, namespace, instanceID), request)(yield)
	}
}

func (c *a2aClient) transport() (a2aclient.Transport, error) {
	connection, err := c.client.grpcConnection()
	if err != nil {
		return nil, err
	}
	return a2agrpc.NewGRPCTransportFromClient(a2apb.NewA2AServiceClient(connection)), nil
}

func routeParams(userID, namespace, instanceID string) a2aclient.ServiceParams {
	params := a2aclient.ServiceParams{
		apia2a.AgentInstanceNamespaceHeader: {namespace},
		apia2a.AgentInstanceIDHeader:        {instanceID},
	}
	if userID != "" {
		params["x-user-id"] = []string{userID}
	}
	return params
}
