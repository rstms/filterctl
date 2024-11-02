package cmd

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestParseFile(t *testing.T) {

	Debug = true
	Verbose = true
	err := InitIdentity()
	require.Nil(t, err)

	input, err := os.Open("testdata/message")
	require.Nil(t, err)

	err = ParseFile(input)
	require.Nil(t, err)

	headerLen := len(Headers)
	require.Greater(t, headerLen, 1)

	subject, ok := Headers["Subject"]
	require.NotNil(t, ok)
	require.NotNil(t, subject)
}
