package cli

import (
	"io"

	clioutput "github.com/kagent-dev/kagent/go/core/cli/internal/cli/output"
)

func printOutput(out io.Writer, format string, data any, tableHeaders []string, tableRows [][]string) error {
	parsed, err := clioutput.Parse(format)
	if err != nil {
		return err
	}
	return clioutput.Write(out, parsed, data, tableHeaders, tableRows, nil)
}
