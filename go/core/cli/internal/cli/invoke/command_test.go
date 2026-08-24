package invoke

import (
	"strings"
	"testing"

	a2atype "github.com/a2aproject/a2a-go/v2/a2a"
	clia2a "github.com/kagent-dev/kagent/go/core/cli/internal/a2a"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadTask(t *testing.T) {
	text, err := readTask("from flag", "", strings.NewReader("unused"))
	require.NoError(t, err)
	assert.Equal(t, "from flag", text)

	text, err = readTask("", "-", strings.NewReader("from stdin"))
	require.NoError(t, err)
	assert.Equal(t, "from stdin", text)

	_, err = readTask("", "", strings.NewReader(""))
	assert.EqualError(t, err, "--task or --file is required")
}

// Assembly itself is covered in internal/a2a.
func TestStreamEventTextRoutesArtifactsThroughTheAssembler(t *testing.T) {
	assembler := clia2a.NewArtifactAssembler()
	update := &a2atype.TaskArtifactUpdateEvent{
		Artifact: &a2atype.Artifact{
			ID:    "response",
			Parts: a2atype.ContentParts{a2atype.NewTextPart("buffered")},
		},
	}

	assert.Empty(t, streamEventText(update, assembler))
	assert.Equal(t, []string{"buffered"}, assembler.Flush())

	message := a2atype.NewMessage(a2atype.MessageRoleAgent, a2atype.NewTextPart("immediate"))
	assert.Equal(t, "immediate", streamEventText(message, assembler))
}

func TestTaskStateError(t *testing.T) {
	tests := []struct {
		state   a2atype.TaskState
		wantErr string
	}{
		{state: a2atype.TaskStateCompleted},
		{state: a2atype.TaskStateWorking},
		{state: a2atype.TaskStateFailed, wantErr: "task ended in state TASK_STATE_FAILED: model unavailable"},
		{state: a2atype.TaskStateRejected, wantErr: "task ended in state TASK_STATE_REJECTED: model unavailable"},
		{state: a2atype.TaskStateCanceled, wantErr: "task ended in state TASK_STATE_CANCELED: model unavailable"},
		{state: a2atype.TaskStateInputRequired, wantErr: "task requires continuation in state TASK_STATE_INPUT_REQUIRED, but continuing the task is not supported: model unavailable"},
		{state: a2atype.TaskStateAuthRequired, wantErr: "task requires continuation in state TASK_STATE_AUTH_REQUIRED, but continuing the task is not supported: model unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			event := &a2atype.TaskStatusUpdateEvent{Status: a2atype.TaskStatus{
				State:   tt.state,
				Message: a2atype.NewMessage(a2atype.MessageRoleAgent, a2atype.NewTextPart("model unavailable")),
			}}
			err := taskStateError(event)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.wantErr)
		})
	}

	assert.EqualError(t, taskStateError(&a2atype.Task{
		Status: a2atype.TaskStatus{State: a2atype.TaskStateFailed},
	}), "task ended in state TASK_STATE_FAILED")
}
