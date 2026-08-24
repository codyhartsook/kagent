package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintOutputUsesRequestedFormat(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, printOutput(&out, "json", []string{}, []string{"VALUE"}, nil))
	assert.JSONEq(t, `[]`, out.String())
}
