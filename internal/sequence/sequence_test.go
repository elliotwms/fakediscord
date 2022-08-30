package sequence

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNext(t *testing.T) {
	for i := 1; i <= 3; i++ {
		require.Equal(t, int64(i), Next())
	}
}
