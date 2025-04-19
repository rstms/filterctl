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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
)

// rescanCmd represents the scan command
var rescanCmd = &cobra.Command{
	Use:   "rescan MESSAGE_FILE",
	Short: "rescan messages with rspamd",
	Long: `
Read folder name or message_ids from MESSAGE_FILE, and rescan designated
messages rspamd, address-books, spam-classes, rewriting message headers.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetDefault("rescand_url", "https://127.0.0.1:2017")
		url := viper.GetString("rescand_url")
		rescan, err := NewAPIClient(url)
		cobra.CheckErr(err)

		var data []byte
		filename := args[0]
		if filename == "" || filename == "-" {
			var buf bytes.Buffer
			_, err := io.Copy(&buf, os.Stdin)
			cobra.CheckErr(err)
			data = buf.Bytes()
		} else {
			data, err = os.ReadFile(filename)
			if err != nil {
				cobra.CheckErr(fmt.Errorf("failed reading message selection file: %v", err))
			}
			if !viper.GetBool("no_remove") {
				err = os.Remove(filename)
				cobra.CheckErr(err)
			}
		}
		var request APIRescanRequest
		err = json.Unmarshal(data, &request)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("failed decoding message selection file: %v", err))
		}
		request.Username = viper.GetString("sender")

		var response APIResponse
		text, err := rescan.Post("/rescan/", &request, &response)
		cobra.CheckErr(err)
		fmt.Println(text)
	},
}

func init() {
	rootCmd.AddCommand(rescanCmd)
}
