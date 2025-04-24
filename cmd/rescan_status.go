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
	"github.com/spf13/viper"
)

var rescanStatusCmd = &cobra.Command{
	Use:   "rescanstatus [ID]",
	Short: "request rescan job status",
	Long: `
Return status of active rescan jobs.  If ID is specified, request status of
a single rescan job, otherwise request status of all active jobs.
`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("rescand_url", "https://127.0.0.1:2017")
		url := viper.GetString("rescand_url")
		rescan, err := NewAPIClient(url)
		cobra.CheckErr(err)

		var text string
		if len(args) == 0 {
			var response APIRescanStatusResponse
			text, err = rescan.Get("/rescan/", &response)
		} else {
			var response APIRescanResponse
			text, err = rescan.Get(fmt.Sprintf("/rescan/%s/", args[0]), &response)
		}
		cobra.CheckErr(err)
		fmt.Println(text)
	},
}

func init() {
	rootCmd.AddCommand(rescanStatusCmd)
}
