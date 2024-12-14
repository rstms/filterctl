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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// usageCmd represents the usage command
var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "describe available commands",
	Long: `
Output a description for each of the commands that may be used on the
Subject line of an email to filterctl@emaildomain.ext.
`,
	Run: func(cmd *cobra.Command, args []string) {
		rule := "------------------------------------------------------------------------------\n"

		type Response struct {
			Success bool
			Message string
			Help    []string
		}

		commands := []struct {
			Name   string
			Args   string
			Detail string
		}{
			{"list", "", listCmd.Long},
			{"delete", "[CLASS ...]", deleteCmd.Long},
			{"reset", "[CLASS=THRESHOLD ...]", resetCmd.Long},
			{"set", "CLASS=THRESHOLD", setCmd.Long},
			{"version", "", versionCmd.Long},
			{"usage", "", "\nOutput this message\n"},
		}

		text := `The rspamd classifier on this mailserver adds an 'X-Spam-Score' header to each
message.  This header value ranges between -100.0 and +100.0, with higher
numbers indicating more spam characteristics.

This rspam-classes filter adds an 'X-Spam-Class' header value based on a list
of class names, each associated with a max score value.  Class names may then
be used for message filtering in the email client.

Each email user may customize the classes and thresholds used for their own
account using this email based command interface.  Commands are executed by
sending a message to 'filterctl@your_domain.com' with the command and
arguments as the 'Subject' line.  The message body is ignored.  A reply
message is sent for each command containing output and status.

Subject Line Commands:
`
		text += rule
		for _, cmd := range commands {
			text += fmt.Sprintf("%s %s\n%s\n", cmd.Name, cmd.Args, cmd.Detail)
			text += rule
		}

		response := Response{
			Success: true,
			Message: fmt.Sprintf("%s usage", viper.GetString("sender")),
			Help:    strings.Split(text, "\n"),
		}
		out, err := json.MarshalIndent(&response, "", "  ")
		cobra.CheckErr(err)
		fmt.Println(string(out))
	},
}

func init() {
	rootCmd.AddCommand(usageCmd)
}
