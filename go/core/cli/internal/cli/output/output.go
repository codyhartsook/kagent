package output

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/jedib0t/go-pretty/v6/table"
)

// Format selects the CLI output representation.
type Format string

const (
	// Table renders human-readable tabular output.
	Table Format = "table"
	// JSON renders machine-readable JSON output.
	JSON Format = "json"
	// Agent renders the structured agent envelope.
	Agent Format = "agent"
)

var formats = []string{string(Table), string(JSON), string(Agent)}

// Parse validates an output format name.
func Parse(value string) (Format, error) {
	format := Format(value)
	if format != Table && format != JSON && format != Agent {
		return "", fmt.Errorf("unknown output format %q (valid: %v)", value, formats)
	}
	return format, nil
}

// Names returns the supported output format names.
func Names() []string {
	return slices.Clone(formats)
}

// Help describes a useful follow-up command for agent output.
type Help struct {
	Description string `json:"description"`
	Command     string `json:"command"`
}

// Envelope is the stable top-level shape for agent output.
type Envelope struct {
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
	Help  []Help `json:"help,omitempty"`
}

// Write renders one non-streaming result.
func Write(out io.Writer, format Format, data any, headers []string, rows [][]string, help []Help) error {
	switch format {
	case Table:
		writer := table.NewWriter()
		writer.SetOutputMirror(out)
		writer.AppendHeader(stringsToRow(headers))
		for _, row := range rows {
			writer.AppendRow(stringsToRow(row))
		}
		writer.Render()
		return nil
	case JSON:
		return writeJSON(out, data, true)
	case Agent:
		return writeJSON(out, Envelope{OK: true, Data: data, Help: help}, true)
	default:
		return fmt.Errorf("unknown output format %q", format)
	}
}

// WriteStream renders one JSONL streaming result.
func WriteStream(out io.Writer, format Format, data any) error {
	if format == Agent {
		data = Envelope{OK: true, Data: data}
	}
	return writeJSON(out, data, false)
}

// WriteError renders an agent-mode error envelope.
func WriteError(out io.Writer, err error) error {
	return writeJSON(out, Envelope{OK: false, Error: err.Error()}, false)
}

func writeJSON(out io.Writer, data any, indent bool) error {
	encoder := json.NewEncoder(out)
	if indent {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}
	return nil
}

func stringsToRow(values []string) table.Row {
	row := make(table.Row, len(values))
	for index := range values {
		row[index] = values[index]
	}
	return row
}
