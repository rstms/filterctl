package cmd

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestInitIdentity(t *testing.T) {
	viper.Set("disable_response", true)
	viper.Set("disable_exec", true)
	viper.Set("verbose", true)
	require.Nil(t, InitIdentity())
	data, err := os.ReadFile("testdata/fqdn")
	require.Nil(t, err)
	fqdn := strings.TrimSpace(string(data))
	require.Equal(t, Hostname, fqdn)
}

func TestParseFile(t *testing.T) {
	TestInitIdentity(t)
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
