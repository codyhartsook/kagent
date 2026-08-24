package client

import (
	"context"
	"net"
	"testing"
	"time"

	a2atype "github.com/a2aproject/a2a-go/v2/a2a"
	a2apb "github.com/a2aproject/a2a-go/v2/a2apb/v1"
	"github.com/a2aproject/a2a-go/v2/a2apb/v1/pbconv"
	apiv1alpha1 "github.com/kagent-dev/kagent/go/api/gen/kagent/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

type recordingAgentInstanceService struct {
	apiv1alpha1.UnimplementedAgentInstanceServiceServer
	metadata metadata.MD
	request  *apiv1alpha1.CreateAgentInstanceRequest
}

func (s *recordingAgentInstanceService) CreateAgentInstance(ctx context.Context, request *apiv1alpha1.CreateAgentInstanceRequest) (*apiv1alpha1.CreateAgentInstanceResponse, error) {
	s.metadata, _ = metadata.FromIncomingContext(ctx)
	s.request = request
	return &apiv1alpha1.CreateAgentInstanceResponse{AgentInstance: testAgentInstance()}, nil
}

type recordingA2AService struct {
	a2apb.UnimplementedA2AServiceServer
	metadata           metadata.MD
	request            *a2apb.SendMessageRequest
	streamHasDeadline  bool
	streamUserMetadata string
}

func (s *recordingA2AService) SendMessage(ctx context.Context, request *a2apb.SendMessageRequest) (*a2apb.SendMessageResponse, error) {
	s.metadata, _ = metadata.FromIncomingContext(ctx)
	s.request = request
	response, err := pbconv.ToProtoSendMessageResponse(&a2atype.Task{
		ID: "task-1", ContextID: testInstanceID,
		Status: a2atype.TaskStatus{State: a2atype.TaskStateCompleted},
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s *recordingA2AService) SendStreamingMessage(request *a2apb.SendMessageRequest, stream grpc.ServerStreamingServer[a2apb.StreamResponse]) error {
	_, s.streamHasDeadline = stream.Context().Deadline()
	md, _ := metadata.FromIncomingContext(stream.Context())
	s.streamUserMetadata = first(md.Get("x-user-id"))
	response, err := pbconv.ToProtoStreamResponse(&a2atype.Task{
		ID: "task-1", ContextID: testInstanceID,
		Status: a2atype.TaskStatus{State: a2atype.TaskStateCompleted},
	})
	if err != nil {
		return err
	}
	return stream.Send(response)
}

const testInstanceID = "d6551fde-f0bb-4d67-b31f-8e2d0150f931"

func TestAgentInstanceAndA2AClientsUsePublicGRPCContracts(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	instances := &recordingAgentInstanceService{}
	a2aService := &recordingA2AService{}
	server := grpc.NewServer()
	apiv1alpha1.RegisterAgentInstanceServiceServer(server, instances)
	a2apb.RegisterA2AServiceServer(server, a2aService)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})

	clientSet := New("http://unused.invalid",
		WithUserID("cli-user"),
		WithGRPCTarget("passthrough:///bufnet"),
		WithGRPCTimeout(5*time.Second),
		WithGRPCDialOptions(grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		})),
	)
	t.Cleanup(func() { require.NoError(t, clientSet.Close()) })

	created, err := clientSet.AgentInstance.CreateAgentInstance(t.Context(), &apiv1alpha1.CreateAgentInstanceRequest{
		Namespace: "kagent", Harness: "kagent", AgentTemplate: "smoke", RequestId: "request-1",
	})
	require.NoError(t, err)
	assert.Equal(t, testInstanceID, created.GetId())
	assert.Equal(t, "cli-user", first(instances.metadata.Get("x-user-id")))
	assert.Equal(t, "request-1", instances.request.GetRequestId())

	message := a2atype.NewMessage(a2atype.MessageRoleUser, a2atype.NewTextPart("hello"))
	message.ContextID = testInstanceID
	result, err := clientSet.A2A.SendMessage(t.Context(), "kagent", testInstanceID, &a2atype.SendMessageRequest{Message: message})
	require.NoError(t, err)
	task, ok := result.(*a2atype.Task)
	require.True(t, ok)
	assert.Equal(t, a2atype.TaskStateCompleted, task.Status.State)
	assert.Equal(t, "cli-user", first(a2aService.metadata.Get("x-user-id")))
	assert.Equal(t, "kagent", first(a2aService.metadata.Get("x-kagent-agent-instance-namespace")))
	assert.Equal(t, testInstanceID, first(a2aService.metadata.Get("x-kagent-agent-instance-id")))
	assert.Equal(t, testInstanceID, a2aService.request.GetMessage().GetContextId())

	var streamEvents int
	for _, streamErr := range clientSet.A2A.SendStreamingMessage(t.Context(), "kagent", testInstanceID, &a2atype.SendMessageRequest{Message: message}) {
		require.NoError(t, streamErr)
		streamEvents++
	}
	assert.Equal(t, 1, streamEvents)
	assert.False(t, a2aService.streamHasDeadline)
	assert.Equal(t, "cli-user", a2aService.streamUserMetadata)

	deadlineContext, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	for _, streamErr := range clientSet.A2A.SendStreamingMessage(deadlineContext, "kagent", testInstanceID, &a2atype.SendMessageRequest{Message: message}) {
		require.NoError(t, streamErr)
	}
	assert.True(t, a2aService.streamHasDeadline)
}

func testAgentInstance() *apiv1alpha1.AgentInstance {
	return &apiv1alpha1.AgentInstance{
		Id: testInstanceID, Namespace: "kagent",
		Harness:       &apiv1alpha1.ResourceReference{Namespace: "kagent", Name: "kagent"},
		AgentTemplate: &apiv1alpha1.ResourceReference{Namespace: "kagent", Name: "smoke"},
		State:         apiv1alpha1.AgentInstanceState_AGENT_INSTANCE_STATE_READY,
	}
}
