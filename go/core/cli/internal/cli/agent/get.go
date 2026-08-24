package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/kagent-dev/kagent/go/api/database"
	api "github.com/kagent-dev/kagent/go/api/httpapi"
	clioutput "github.com/kagent-dev/kagent/go/core/cli/internal/cli/output"
	"github.com/kagent-dev/kagent/go/core/cli/internal/config"
	"github.com/kagent-dev/kagent/go/core/internal/utils"
)

func GetAgentCmd(cfg *config.Config, resourceName string) {
	client := cfg.Client()

	if resourceName == "" {
		agentList, err := client.Agent.ListAgents(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get agents: %v\n", err)
			return
		}

		if len(agentList.Data) == 0 && cfg.OutputFormat == string(clioutput.Table) {
			fmt.Println("No agents found")
			return
		}

		if err := printAgents(os.Stdout, cfg.OutputFormat, agentList.Data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to print agents: %v\n", err)
			return
		}
	} else {
		agent, err := client.Agent.GetAgent(context.Background(), resourceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get agent %s: %v\n", resourceName, err)
			return
		}
		byt, _ := json.MarshalIndent(agent, "", "  ")
		fmt.Fprintln(os.Stdout, string(byt))
	}
}

func GetSessionCmd(cfg *config.Config, resourceName string) {
	client := cfg.Client()
	if resourceName == "" {
		sessionList, err := client.Session.ListSessions(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get sessions: %v\n", err)
			return
		}

		if len(sessionList.Data) == 0 && cfg.OutputFormat == string(clioutput.Table) {
			fmt.Println("No sessions found")
			return
		}

		if err := printSessions(os.Stdout, cfg.OutputFormat, sessionList.Data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to print sessions: %v\n", err)
			return
		}
	} else {
		session, err := client.Session.GetSession(context.Background(), resourceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get session %s: %v\n", resourceName, err)
			return
		}
		byt, _ := json.MarshalIndent(session, "", "  ")
		fmt.Fprintln(os.Stdout, string(byt))
	}
}

func GetToolCmd(cfg *config.Config) {
	client := cfg.Client()
	toolList, err := client.Tool.ListTools(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get tools: %v\n", err)
		return
	}
	if err := printTools(os.Stdout, cfg.OutputFormat, toolList); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to print tools: %v\n", err)
		return
	}
}

func printTools(out io.Writer, format string, tools []database.Tool) error {
	headers := []string{"#", "NAME", "SERVER_NAME", "DESCRIPTION", "CREATED"}
	rows := make([][]string, len(tools))
	for i, tool := range tools {
		rows[i] = []string{
			strconv.Itoa(i + 1),
			tool.ID,
			tool.ServerName,
			tool.Description,
			tool.CreatedAt.Format(time.RFC3339),
		}
	}

	return printOutput(out, format, tools, headers, rows)
}

func printAgents(out io.Writer, format string, agents []api.AgentResponse) error {
	// Prepare table data
	headers := []string{"#", "NAME", "CREATED", "DEPLOYMENT_READY", "ACCEPTED"}
	rows := make([][]string, len(agents))
	for i, agent := range agents {
		rows[i] = []string{
			strconv.Itoa(i + 1),
			utils.ResourceRefString(agent.Agent.Metadata.Namespace, agent.Agent.Metadata.Name),
			agent.Agent.Metadata.CreationTimestamp.Format(time.RFC3339),
			strconv.FormatBool(agent.Ready),
			strconv.FormatBool(agent.Accepted),
		}
	}

	return printOutput(out, format, agents, headers, rows)
}

func printSessions(out io.Writer, format string, sessions []*database.Session) error {
	headers := []string{"#", "ID", "NAME", "AGENT", "CREATED"}
	rows := make([][]string, len(sessions))
	for i, session := range sessions {
		agentID := ""
		if session.AgentID != nil {
			agentID = *session.AgentID
		}
		sessionName := ""
		if session.Name != nil {
			sessionName = *session.Name
		}
		rows[i] = []string{
			strconv.Itoa(i + 1),
			session.ID,
			sessionName,
			agentID,
			session.CreatedAt.Format(time.RFC3339),
		}
	}

	return printOutput(out, format, sessions, headers, rows)
}
