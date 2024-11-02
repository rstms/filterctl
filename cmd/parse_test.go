package cmd

import (
    "os"
    "testing"
	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {

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
