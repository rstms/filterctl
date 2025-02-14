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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addCmd represents the mkaddr command
var addCmd = &cobra.Command{
	Use:   "add BOOK_NAME EMAIL_ADDRESS",
	Short: "add email address to book",
	Long: `
Add an email address to the named address book
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		filterctld := initAPI()
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
		text, err := filterctld.Post("/filterctl/address/", &request, &response)
		cobra.CheckErr(err)
		if strings.Contains(response.Message, "QueryAddressBook failed: 404 Not Found") {
			_, err := AddAddressBook(request.Username, request.Bookname, "")
			cobra.CheckErr(err)
			text, err = filterctld.Post("/filterctl/address/", &request, &response)
			cobra.CheckErr(err)
		}
		fmt.Println(text)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
