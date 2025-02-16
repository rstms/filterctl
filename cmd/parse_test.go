package cmd

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var testFQDN = "phobos.rstms.net"
var testDomains = []string{"rstms.net"}

func TestInitIdentity(t *testing.T) {
	viper.Set("disable_response", true)
	viper.Set("disable_exec", true)
	viper.Set("verbose", true)
	viper.Set("hostname", testFQDN)
	viper.Set("domains", testDomains)
	require.Nil(t, InitIdentity())
	require.Equal(t, Hostname, testFQDN)
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
