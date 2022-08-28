package sequence

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNext(t *testing.T) {
	for i := 1; i <= 3; i++ {
		require.Equal(t, int64(i), Next())
	}
}
