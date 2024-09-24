package jessy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	s := getTestStruct()
	h1, err := Hash(s)
	require.NoError(t, err)

	s = getTestStruct()
	h2, err := Hash(s)
	require.NoError(t, err)

	s = getTestStruct()
	h3, err := Hash(&s)
	require.NoError(t, err)

	require.Equal(t, h1, h2)
	require.Equal(t, h1, h3)
}
