package snowflake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateBeforeConfigure(t *testing.T) {
	node = nil // ensure node is not set up for this test

	require.PanicsWithValue(t, "snowflake.Generate called before snowflake.Configure", func() {
		Generate()
	})
}

func TestGenerate(t *testing.T) {
	require.NoError(t, Configure(0))
	require.NotEmpty(t, Generate())
	t.Logf(Generate().String())
}
