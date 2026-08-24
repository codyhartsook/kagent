package invoke

import (
	"errors"
	"fmt"
	"io"
	"os"

	a2atype "github.com/a2aproject/a2a-go/v2/a2a"
	clia2a "github.com/kagent-dev/kagent/go/core/cli/internal/a2a"
	legacycli "github.com/kagent-dev/kagent/go/core/cli/internal/cli/agent"
	clioutput "github.com/kagent-dev/kagent/go/core/cli/internal/cli/output"
	"github.com/kagent-dev/kagent/go/core/cli/internal/config"
	"github.com/spf13/cobra"
)

// NewCommand creates the AgentInstance invoke command.
func NewCommand(cfg *config.Config) *cobra.Command {
	var instanceID, task, file string
	var stream bool
	cmd := &cobra.Command{
		Use:          "invoke",
		Short:        "Invoke an AgentInstance",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			text, err := readTask(task, file, cmd.InOrStdin())
			if err != nil {
				return err
			}
			format, err := clioutput.Parse(cfg.OutputFormat)
			if err != nil {
				return err
			}
			client, cleanup, err := legacycli.Connect(cmd.Context(), cfg)
			if err != nil {
				return err
			}
			defer cleanup()
			message := a2atype.NewMessage(a2atype.MessageRoleUser, a2atype.NewTextPart(text))
			message.ContextID = instanceID
			request := &a2atype.SendMessageRequest{Message: message}
			if !stream {
				result, err := client.A2A.SendMessage(cmd.Context(), cfg.Namespace, instanceID, request)
				if err != nil {
					return fmt.Errorf("invoke AgentInstance %s: %w", instanceID, err)
				}
				if err := taskStateError(result); err != nil {
					return fmt.Errorf("invoke AgentInstance %s: %w", instanceID, err)
				}
				if format == clioutput.Table {
					_, err = fmt.Fprintln(cmd.OutOrStdout(), clia2a.AgentText(result))
					return err
				}
				return clioutput.Write(cmd.OutOrStdout(), format, result, nil, nil, nil)
			}
			assembler := clia2a.NewArtifactAssembler()
			for event, eventErr := range client.A2A.SendStreamingMessage(cmd.Context(), cfg.Namespace, instanceID, request) {
				if eventErr != nil {
					return fmt.Errorf("invoke AgentInstance %s: %w", instanceID, eventErr)
				}
				if err := taskStateError(event); err != nil {
					return fmt.Errorf("invoke AgentInstance %s: %w", instanceID, err)
				}
				if format == clioutput.Table {
					if text := streamEventText(event, assembler); text != "" {
						if _, err := fmt.Fprint(cmd.OutOrStdout(), text); err != nil {
							return err
						}
					}
				} else if err := clioutput.WriteStream(cmd.OutOrStdout(), format, event); err != nil {
					return err
				}
			}
			if format == clioutput.Table {
				for _, text := range assembler.Flush() {
					if _, err := fmt.Fprint(cmd.OutOrStdout(), text); err != nil {
						return err
					}
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout())
			}
			return err
		},
	}
	cmd.Flags().StringVar(&instanceID, "agent-instance", "", "AgentInstance ID")
	cmd.Flags().StringVarP(&task, "task", "t", "", "Task text")
	cmd.Flags().StringVarP(&file, "file", "f", "", "Read task text from a file or - for stdin")
	cmd.Flags().BoolVarP(&stream, "stream", "S", false, "Stream the response")
	_ = cmd.MarkFlagRequired("agent-instance")
	cmd.MarkFlagsMutuallyExclusive("task", "file")
	return cmd
}

func streamEventText(event any, assembler *clia2a.ArtifactAssembler) string {
	if update, ok := event.(*a2atype.TaskArtifactUpdateEvent); ok {
		text, _ := assembler.Add(update)
		return text
	}
	return clia2a.AgentText(event)
}

func taskStateError(event any) error {
	var status a2atype.TaskStatus
	switch value := event.(type) {
	case *a2atype.Task:
		status = value.Status
	case *a2atype.TaskStatusUpdateEvent:
		status = value.Status
	default:
		return nil
	}

	state := status.State
	continuationUnsupported := false
	switch state {
	case a2atype.TaskStateInputRequired, a2atype.TaskStateAuthRequired:
		continuationUnsupported = true
	case a2atype.TaskStateCanceled, a2atype.TaskStateFailed, a2atype.TaskStateRejected:
	default:
		return nil
	}
	message := fmt.Sprintf("task ended in state %s", state)
	if continuationUnsupported {
		message = fmt.Sprintf("task requires continuation in state %s, but continuing the task is not supported", state)
	}
	if detail := clia2a.AgentText(status.Message); detail != "" {
		message += ": " + detail
	}
	return errors.New(message)
}

func readTask(task, file string, stdin io.Reader) (string, error) {
	if task != "" {
		return task, nil
	}
	if file == "" {
		return "", fmt.Errorf("--task or --file is required")
	}
	var data []byte
	var err error
	if file == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(file)
	}
	if err != nil {
		return "", fmt.Errorf("read task: %w", err)
	}
	return string(data), nil
}
