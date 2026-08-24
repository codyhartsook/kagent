package output

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatsAndAgentError(t *testing.T) {
	for _, value := range Names() {
		_, err := Parse(value)
		require.NoError(t, err)
	}
	_, err := Parse("yaml")
	assert.ErrorContains(t, err, "valid: [table json agent]")

	var out bytes.Buffer
	require.NoError(t, WriteError(&out, errors.New("not found")))
	assert.JSONEq(t, `{"ok":false,"error":"not found"}`, out.String())
}
