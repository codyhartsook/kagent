package a2a

import (
	"strings"

	a2atype "github.com/a2aproject/a2a-go/v2/a2a"
)

// ArtifactAssembler buffers streamed artifact text until its last chunk.
// Updates carry either an append or a full replacement, so a consumer that
// cannot retract rendered output has to wait for the authoritative chunk.
type ArtifactAssembler struct {
	text  map[a2atype.ArtifactID]string
	order []a2atype.ArtifactID
}

// NewArtifactAssembler creates an empty assembler.
func NewArtifactAssembler() *ArtifactAssembler {
	return &ArtifactAssembler{text: make(map[a2atype.ArtifactID]string)}
}

// Add records an update and returns the assembled text once it is complete.
func (a *ArtifactAssembler) Add(update *a2atype.TaskArtifactUpdateEvent) (string, bool) {
	if update == nil || update.Artifact == nil {
		return "", false
	}
	id := update.Artifact.ID
	text := PartsText(update.Artifact.Parts)
	if update.Append {
		text = a.text[id] + text
	}
	if _, tracked := a.text[id]; !tracked {
		a.order = append(a.order, id)
	}
	a.text[id] = text
	if !update.LastChunk {
		return "", false
	}
	a.drop(id)
	return text, true
}

// Flush drains artifacts that never reached a last chunk, first seen first, so
// a stream that ends early does not discard them.
func (a *ArtifactAssembler) Flush() []string {
	texts := make([]string, 0, len(a.order))
	for _, id := range a.order {
		texts = append(texts, a.text[id])
	}
	a.text = make(map[a2atype.ArtifactID]string)
	a.order = nil
	return texts
}

// Reset discards all buffered artifacts without reporting them.
func (a *ArtifactAssembler) Reset() {
	a.text = make(map[a2atype.ArtifactID]string)
	a.order = nil
}

func (a *ArtifactAssembler) drop(id a2atype.ArtifactID) {
	delete(a.text, id)
	for index, pending := range a.order {
		if pending == id {
			a.order = append(a.order[:index], a.order[index+1:]...)
			break
		}
	}
}

// PartsText concatenates the text parts of a message or artifact.
func PartsText(parts a2atype.ContentParts) string {
	var text strings.Builder
	for _, part := range parts {
		if part == nil {
			continue
		}
		text.WriteString(part.Text())
	}
	return text.String()
}

// AgentText returns the agent-authored text carried by an A2A event or message.
func AgentText(event any) string {
	var parts a2atype.ContentParts
	appendAgent := func(message *a2atype.Message) {
		if message != nil && message.Role == a2atype.MessageRoleAgent {
			parts = append(parts, message.Parts...)
		}
	}
	switch value := event.(type) {
	case *a2atype.Message:
		appendAgent(value)
	case *a2atype.Task:
		appendAgent(value.Status.Message)
		for _, artifact := range value.Artifacts {
			parts = append(parts, artifact.Parts...)
		}
	case *a2atype.TaskStatusUpdateEvent:
		appendAgent(value.Status.Message)
	case *a2atype.TaskArtifactUpdateEvent:
		if value.Artifact != nil {
			parts = append(parts, value.Artifact.Parts...)
		}
	}
	return PartsText(parts)
}
