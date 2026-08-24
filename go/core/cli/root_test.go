package cli_test

import (
	"testing"

	"github.com/kagent-dev/kagent/go/core/cli"
	"github.com/kagent-dev/kagent/go/core/cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewAddsAndDisablesCommands(t *testing.T) {
	root := cli.New(cli.Config{
		Runtime:       &config.Config{OutputFormat: "table"},
		ExtraCommands: []*cobra.Command{{Use: "extra"}, {Use: "kept"}},
		Disabled:      []string{"get session", "extra"},
	})

	assert.Nil(t, child(root, "extra"))
	assert.NotNil(t, child(root, "kept"))
	assert.Nil(t, child(child(root, "get"), "session"))
	assert.NotNil(t, child(child(root, "get"), "agent"))
}

func child(parent *cobra.Command, name string) *cobra.Command {
	for _, command := range parent.Commands() {
		if command.Name() == name {
			return command
		}
	}
	return nil
}
