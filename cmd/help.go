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

func rule() {
	fmt.Println("------------------------------------------------------------------------------")
}

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "describe available commands",
	Long: `
Output a description for each of the commands that may be used on the
Subject line of an email to filterctl@emaildomain.ext.
`,
	Run: func(cmd *cobra.Command, args []string) {
		commands := []struct {
			Name   string
			Args   string
			Detail string
		}{
			{"list", "", listCmd.Long},
			{"delete", "[CLASS, ...]", deleteCmd.Long},
			{"reset", "[CLASS=THRESHOLD, ...]", resetCmd.Long},
			{"set", "CLASS=THRESHOLD", setCmd.Long},
			{"version", "", versionCmd.Long},
			{"help", "", "\nOutput this message\n"},
		}
		fmt.Println(`
The rspamd classifier on this mailserver adds an 'X-Spam-Score' header to each
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
`)
		rule()
		for _, cmd := range commands {
			fmt.Printf("%s %s\n%s", cmd.Name, cmd.Args, cmd.Detail)
			rule()
		}
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// helpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// helpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
