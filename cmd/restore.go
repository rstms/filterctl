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
	"github.com/rstms/mabctl/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var noRemove bool

var restoreCmd = &cobra.Command{
	Use:   "restore [RESTORE_FILE]",
	Short: "restore carddav config",
	Long: `
Restore the cardDAV config for the sender user from the JSON data in RESTORE_FILE.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		var file *os.File
		var err error
		if filename == "" || filename == "-" {
			file = os.Stdin
		} else {
			file, err = os.Open(filename)
			cobra.CheckErr(err)
			if !noRemove {
				defer func() {
					err := os.Remove(filename)
					cobra.CheckErr(err)
				}()
			}
			defer file.Close()
		}
		MAB := InitAPI()
		var response APIResponse
		var request APIRestoreRequest
		request.Username = viper.GetString("sender")
		request.Dump = api.ConfigDump{}
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&request.Dump)
		cobra.CheckErr(err)
		text, err := MAB.Post("/filterctl/restore/", &request, &response)
		cobra.CheckErr(err)
		fmt.Println(text)
	},
}

func init() {
	restoreCmd.Flags().BoolVar(&noRemove, "no-remove", false, "disable deletion of input file")
	rootCmd.AddCommand(restoreCmd)
}
