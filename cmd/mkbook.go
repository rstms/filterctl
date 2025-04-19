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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mkbookCmd = &cobra.Command{
	Use:   "mkbook BOOK_NAME [DESCRIPTION]",
	Short: "create a new address book",
	Long: `
Create a new address book under the sender's address with the NAME and
DESCRIPTION.  Returns a data structure including the new book token and URI
`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		bookName := args[0]
		description := bookName
		if len(args) > 1 {
			description = args[1]
		}
		filterctl := NewFilterctlClient()
		text, err := AddAddressBook(filterctl, viper.GetString("sender"), bookName, description)
		cobra.CheckErr(err)
		fmt.Println(text)
	},
}

func init() {
	rootCmd.AddCommand(mkbookCmd)
}

func AddAddressBook(filterctl *APIClient, username, bookname, description string) (string, error) {
	type Request struct {
		Username    string
		Bookname    string
		Description string
	}
	request := Request{
		Username:    username,
		Bookname:    bookname,
		Description: description,
	}
	var response APIResponse
	result, err := filterctl.Post("/filterctl/book/", &request, &response)
	if err != nil {
		return "", err
	}
	log.Printf("AddAddressBook: %s\n", result)
	return result, nil

}
