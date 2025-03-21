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
	"os"

	"github.com/rstms/mabctl/api"
	"github.com/rstms/rspamd-classes/classes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "display version",
	Long: `
Outputs program name, version, rspamd_classes library version, uid, and gid.
`,
	Run: func(cmd *cobra.Command, args []string) {

		var response APIVersionResponse
		var err error

		response.User = viper.GetString("sender")
		response.Request, err = DecodedMessageID(viper.GetString("message_id"))
		cobra.CheckErr(err)
		response.Success = true
		response.Message = fmt.Sprintf("%s version", viper.GetString("sender"))
		response.Name = os.Args[0]
		response.Version = Version
		response.Classes = classes.Version
		response.Mabctl = api.Version
		response.UID = os.Getuid()
		response.GID = os.Getgid()

		out, err := json.MarshalIndent(&response, "", "  ")
		cobra.CheckErr(err)
		fmt.Println(string(out))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
