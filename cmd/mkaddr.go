/*
Copyright © 2024 Matt Krueger <mkrueger@rstms.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mkaddrCmd = &cobra.Command{
	Use:   "mkaddr BOOK_NAME EMAIL_ADDRESS",
	Short: "add email address to book",
	Long: `
Add an email address to the named address book
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		api := InitAPI()
		type Request struct {
			Username string
			Bookname string
			Address  string
			Name     string
		}
		request := Request{
			Username: viper.GetString("sender"),
			Bookname: args[0],
			Address:  args[1],
		}
		var response APIResponse
		for {
			text, err := api.Post("/filterctl/address/", &request, &response)
			cobra.CheckErr(err)
			switch {
			case strings.Contains(response.Message, "AddAddress failed: Unknown user:"):
				_, err := AddUser(api, request.Username, "", "")
				cobra.CheckErr(err)
			case strings.Contains(response.Message, "QueryAddressBook failed: 404 Not Found"):
				_, err := AddAddressBook(api, request.Username, request.Bookname, "")
				cobra.CheckErr(err)
			default:
				fmt.Println(text)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(mkaddrCmd)
}

func AddUser(api *APIClient, username, email, password string) (string, error) {
	type Request struct {
		Username string
		Email    string
		Password string
	}
	request := Request{
		Username: username,
		Email:    email,
		Password: password,
	}
	var response APIResponse
	result, err := api.Post("/filterctl/user/", &request, &response)
	if err != nil {
		return "", err
	}
	log.Printf("AddUser: %s\n", result)
	return result, nil
}
