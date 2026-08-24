package agentinstance

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	apiv1alpha1 "github.com/kagent-dev/kagent/go/api/gen/kagent/api/v1alpha1"
	legacycli "github.com/kagent-dev/kagent/go/core/cli/internal/cli/agent"
	clioutput "github.com/kagent-dev/kagent/go/core/cli/internal/cli/output"
	"github.com/kagent-dev/kagent/go/core/cli/internal/config"
	"github.com/spf13/cobra"
)

// NewCreateCommand creates the AgentInstance create command.
func NewCreateCommand(cfg *config.Config) *cobra.Command {
	var harness, template, requestID string
	cmd := &cobra.Command{
		Use:          "agent-instance",
		Short:        "Create an AgentInstance",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if requestID == "" {
				requestID = uuid.NewString()
			}
			client, cleanup, err := legacycli.Connect(cmd.Context(), cfg)
			if err != nil {
				return err
			}
			defer cleanup()
			instance, err := client.AgentInstance.CreateAgentInstance(cmd.Context(), &apiv1alpha1.CreateAgentInstanceRequest{
				Namespace: cfg.Namespace, Harness: harness, AgentTemplate: template, RequestId: requestID,
			})
			if err != nil {
				return fmt.Errorf("create AgentInstance: %w", err)
			}
			if instance.GetState() != apiv1alpha1.AgentInstanceState_AGENT_INSTANCE_STATE_READY {
				return fmt.Errorf("create AgentInstance returned state %s, want READY", instance.GetState())
			}
			format, err := clioutput.Parse(cfg.OutputFormat)
			if err != nil {
				return err
			}
			return write(cmd, format, instance, []clioutput.Help{{
				Description: "Invoke this AgentInstance",
				Command:     fmt.Sprintf("kagent invoke --agent-instance %s --task MESSAGE", instance.GetId()),
			}})
		},
	}
	cmd.Flags().StringVar(&harness, "harness", "", "Harness name")
	cmd.Flags().StringVar(&template, "agent-template", "", "AgentTemplate name")
	cmd.Flags().StringVar(&requestID, "request-id", "", "Idempotency key (defaults to a generated UUID)")
	_ = cmd.MarkFlagRequired("harness")
	_ = cmd.MarkFlagRequired("agent-template")
	return cmd
}

// NewGetCommand creates the AgentInstance get and list command.
func NewGetCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:          "agent-instance [id]",
		Short:        "Get an AgentInstance or list AgentInstances",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, cleanup, err := legacycli.Connect(cmd.Context(), cfg)
			if err != nil {
				return err
			}
			defer cleanup()
			format, err := clioutput.Parse(cfg.OutputFormat)
			if err != nil {
				return err
			}
			if len(args) == 1 {
				instance, err := client.AgentInstance.GetAgentInstance(cmd.Context(), &apiv1alpha1.GetAgentInstanceRequest{
					Namespace: cfg.Namespace, AgentInstanceId: args[0],
				})
				if err != nil {
					return fmt.Errorf("get AgentInstance %s: %w", args[0], err)
				}
				return write(cmd, format, instance, nil)
			}
			response, err := client.AgentInstance.ListAgentInstances(cmd.Context(), &apiv1alpha1.ListAgentInstancesRequest{Namespace: cfg.Namespace})
			if err != nil {
				return fmt.Errorf("list AgentInstances: %w", err)
			}
			return writeList(cmd, format, response)
		},
	}
}

func write(cmd *cobra.Command, format clioutput.Format, instance *apiv1alpha1.AgentInstance, help []clioutput.Help) error {
	return clioutput.Write(cmd.OutOrStdout(), format, instance,
		[]string{"ID", "TEMPLATE", "HARNESS", "STATE", "CREATED"},
		[][]string{instanceRow(instance)}, help)
}

func writeList(cmd *cobra.Command, format clioutput.Format, response *apiv1alpha1.ListAgentInstancesResponse) error {
	rows := make([][]string, 0, len(response.GetAgentInstances()))
	for _, instance := range response.GetAgentInstances() {
		rows = append(rows, instanceRow(instance))
	}
	return clioutput.Write(cmd.OutOrStdout(), format, response,
		[]string{"ID", "TEMPLATE", "HARNESS", "STATE", "CREATED"}, rows, nil)
}

func instanceRow(instance *apiv1alpha1.AgentInstance) []string {
	created := ""
	if instance.GetCreatedAt() != nil {
		created = instance.GetCreatedAt().AsTime().Format(time.RFC3339)
	}
	return []string{
		instance.GetId(), instance.GetAgentTemplate().GetName(), instance.GetHarness().GetName(),
		instance.GetState().String(), created,
	}
}
