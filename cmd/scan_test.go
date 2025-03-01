package cmd

import (
	"fmt"
	"github.com/rstms/mabctl/api"
	"github.com/stretchr/testify/require"
	//"github.com/spf13/viper"
	"testing"
)

func TestScanCommand(t *testing.T) {
	//err := InitIdentity()
	//require.Nil(t, err)
	filterctld := InitAPI()
	sender := "sender@example.org"
	address := "address@example.org"
	var response api.BooksResponse
	path := fmt.Sprintf("/filterctl/scan/%s/%s/", sender, address)
	text, err := filterctld.Get(path, &response)
	require.Nil(t, err)
	fmt.Printf("text=%v\n", text)
	fmt.Printf("response=%v\n", response)
}
