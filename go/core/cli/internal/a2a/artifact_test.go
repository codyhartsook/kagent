package a2a

import (
	"testing"

	a2atype "github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/stretchr/testify/assert"
)

func artifactUpdate(text string, appendChunk, lastChunk bool) *a2atype.TaskArtifactUpdateEvent {
	return &a2atype.TaskArtifactUpdateEvent{
		Append:    appendChunk,
		LastChunk: lastChunk,
		Artifact: &a2atype.Artifact{
			ID:    "response",
			Parts: a2atype.ContentParts{a2atype.NewTextPart(text)},
		},
	}
}

func assertAdd(t *testing.T, assembler *ArtifactAssembler, update *a2atype.TaskArtifactUpdateEvent, wantText string, wantDone bool) {
	t.Helper()
	text, done := assembler.Add(update)
	assert.Equal(t, wantText, text)
	assert.Equal(t, wantDone, done)
}

func TestArtifactAssemblerReportsOnlyTheLastChunk(t *testing.T) {
	assembler := NewArtifactAssembler()

	assertAdd(t, assembler, artifactUpdate("the", false, false), "", false)
	assertAdd(t, assembler, artifactUpdate("re", true, false), "", false)
	assertAdd(t, assembler, artifactUpdate("there", false, true), "there", true)
	assert.Empty(t, assembler.Flush())
}

func TestArtifactAssemblerFlushesTruncatedArtifactsInOrder(t *testing.T) {
	assembler := NewArtifactAssembler()
	second := artifactUpdate("second", false, false)
	second.Artifact.ID = "other"

	assertAdd(t, assembler, artifactUpdate("first", false, false), "", false)
	assertAdd(t, assembler, second, "", false)

	assert.Equal(t, []string{"first", "second"}, assembler.Flush())
	assert.Empty(t, assembler.Flush())
}

func TestArtifactAssemblerResetDiscardsBufferedText(t *testing.T) {
	assembler := NewArtifactAssembler()

	assertAdd(t, assembler, artifactUpdate("partial", false, false), "", false)
	assembler.Reset()

	assert.Empty(t, assembler.Flush())
}

func TestArtifactAssemblerIgnoresEmptyUpdates(t *testing.T) {
	assembler := NewArtifactAssembler()

	assertAdd(t, assembler, nil, "", false)
	assertAdd(t, assembler, &a2atype.TaskArtifactUpdateEvent{}, "", false)
}

func TestAgentTextReturnsOnlyAgentAuthoredText(t *testing.T) {
	assert.Empty(t, AgentText(a2atype.NewMessage(a2atype.MessageRoleUser, a2atype.NewTextPart("prompt"))))
	assert.Equal(t, "reply", AgentText(a2atype.NewMessage(a2atype.MessageRoleAgent, a2atype.NewTextPart("reply"))))
	assert.Equal(t, "artifact", AgentText(&a2atype.TaskArtifactUpdateEvent{
		Artifact: &a2atype.Artifact{Parts: a2atype.ContentParts{a2atype.NewTextPart("artifact")}},
	}))
	assert.Empty(t, AgentText(&a2atype.TaskStatusUpdateEvent{
		Status: a2atype.TaskStatus{State: a2atype.TaskStateWorking},
	}))
}

func TestAgentTextReadsTaskSnapshotStatusAndArtifacts(t *testing.T) {
	task := &a2atype.Task{
		Status: a2atype.TaskStatus{
			Message: a2atype.NewMessage(a2atype.MessageRoleAgent, a2atype.NewTextPart("status ")),
		},
		Artifacts: []*a2atype.Artifact{
			{Parts: a2atype.ContentParts{a2atype.NewTextPart("result")}},
		},
	}

	assert.Equal(t, "status result", AgentText(task))
}
