package cmd

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var testFQDN = "phobos.rstms.net"
var testDomains = []string{"rstms.net"}

func configure(t *testing.T) {
	viper.SetConfigType("yaml")
	viper.AddConfigPath("testdata")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	require.Nil(t, err)
	viper.Set("disable_response", true)
	viper.Set("disable_exec", true)
	viper.Set("verbose", true)
	viper.Set("hostname", testFQDN)
	viper.Set("domains", testDomains)
	require.Nil(t, InitIdentity())
	require.Equal(t, Hostname, testFQDN)
}

func TestParseFile(t *testing.T) {
	configure(t)

	viper.Set("sender", "mkrueger@rstms.net")
	input, err := os.Open("testdata/message")
	require.Nil(t, err)

	err = ParseFile(input)
	require.Nil(t, err)
}
