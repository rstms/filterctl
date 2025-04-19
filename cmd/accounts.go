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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts RESTORE_FILE",
	Short: "cardDAV user accounts",
	Long: `
Read a JSON list of user email addresses from REQUEST_FILE.  Output cardDAV
credentials for each address.  When used with the email subject command the
message body must contain the JSON email address list.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		var err error
		var file *os.File

		if filename == "" || filename == "-" {
			file = os.Stdin
		} else {
			file, err = os.Open(filename)
			cobra.CheckErr(err)
			if !viper.GetBool("no_remove") {
				defer func() {
					err := os.Remove(filename)
					cobra.CheckErr(err)
				}()
			}
			defer file.Close()
		}
		filterctl := NewFilterctlClient()
		decoder := json.NewDecoder(file)
		users := []string{}
		err = decoder.Decode(&users)
		cobra.CheckErr(err)
		var table APIAccountsResponse
		table.Accounts = make(map[string]string)
		table.User = viper.GetString("sender")
		table.Request = "accounts query"
		table.Message = "cardDAV user accounts"
		table.Success = true
		for _, user := range users {
			var response APIPasswordResponse
			path := fmt.Sprintf("/filterctl/passwd/%s/", user)
			viper.Set("sender", user)
			_, err := filterctl.Get(path, &response)
			cobra.CheckErr(err)
			table.Accounts[response.User] = response.Password
		}
		text, err := json.MarshalIndent(&table, "", "  ")
		cobra.CheckErr(err)
		fmt.Println(string(text))
	},
}

func init() {
	rootCmd.AddCommand(accountsCmd)
}
