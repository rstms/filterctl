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

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete ADDRESS [CLASS, ...]",
	Short: "delete all or selected rspamd classes",
	Long: `
Delete rspamd filter classes for an email ADDRESS.  If one or more optional
CLASS names are provided, delete the listed CLASSES from the configuration.
Without CLASS names, delete the entire configuration for the address.
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		address := args[0]
		var response string
		if len(args) == 1 {
			path := fmt.Sprintf("/filterctl/classes/%s", address)
			r, err := api.Delete(path)
			cobra.CheckErr(err)
			response = r
		} else {
			for _, class := range args[1:] {
				path := fmt.Sprintf("/filterctl/classes/%s/%s", address, class)
				r, err := api.Delete(path)
				cobra.CheckErr(err)
				response = r
			}
		}
		fmt.Println(response)
	},
}

func init() {
	classCmd.AddCommand(deleteCmd)
}
