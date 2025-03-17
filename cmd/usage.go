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

		commands := []struct {
			Name   string
			Args   string
			Detail string
		}{
			{"classes", "", classesCmd.Long},
			{"set", "CLASS=THRESHOLD", setCmd.Long},
			{"delete", "[CLASS ...]", deleteCmd.Long},
			{"reset", "[CLASS=THRESHOLD ...]", resetCmd.Long},
			{"classify", "SCORE", classifyCmd.Long},
			{"books", "", booksCmd.Long},
			{"addrs", "BOOK_NAME", addrsCmd.Long},
			{"mkbook", "BOOK_NAME [DESCRIPTION]", mkbookCmd.Long},
			{"rmbook", "BOOK_NAME", rmbookCmd.Long},
			{"mkaddr", "BOOK_NAME EMAIL_ADDRESS", mkaddrCmd.Long},
			{"rmaddr", "BOOK_NAME EMAIL_ADDRESS", rmaddrCmd.Long},
			{"scan", "EMAIL_ADDRESS", scanCmd.Long},
			{"passwd", "", passwdCmd.Long},
			{"dump", "", dumpCmd.Long},
			{"restore", "", restoreCmd.Long},
			{"version", "", versionCmd.Long},
			{"usage", "", "\nOutput this message\n"},
		}

		help := `### Mail Filter Control ####
# Filter Control Address #
The email address 'filterctl@[account_domain]' accepts messages only from
internal users connecting on a TLS-secured authorized connection.  Messages
may be sent to this address to examine or modify the configuration of several
filter mechanisms. 

In this document [account_domain] represents the full domain name for any 
email account. The filter control address for a user with the email address 
'mailuser@mailserver.com' would be 'filterctl@mailserver.com'

Each email user may customize parameters and settings used for their account
with this email-based command interface.  Commands are executed by sending a
message to 'filterctl@[account_domain]' with the command and any arguments as
the 'Subject' line.  The message body is usually ignored.  (see below for an
exception to this rule)  When the command is executed by the mailserver a
response message is sent from 'filterctl@[account_domain]' with the subject
'filterctl response'.  The body of the response message contains output from
the command.

# X-Spam-Score Header # 
The rspamd classifier on the mailserver adds an 'X-Spam-Score' header to each
incoming message.  This header value generally ranges between -20.0 and +20.0,
with higher numbers indicating more spam characteristics.

# X-Spam-Class Header #
To facilitate the use of filter rules in the email client, The spam classes
filter adds an 'X-Spam-Class' header value based on a list of class names.
Each class is associated with a maximum score value.  The highest class is
'spam' with a fixed maximum.  A default set of classes is used if the user
has not set any custom classes.

# Address Book Filter #
The system maintains address books which may be used to classify mail by
sender address bypassing analysis of message content.  These address books are
stored on a remote CardDAV server.  Note that the address book filter's
address books are separate from the mail client's address books and are used
only for filtering inbound mail.

# X-Address-Book Header #
The address book filter adds an 'X-Address-Book' header value to any incoming
message with a 'From' address that is listed in any of the address books
associated with a recipient email address.  The header's value is set to the
name of the address book containing the sender address.  Multiple headers may
be present if a sender address is listed in multiple filter address books.

# Plus Extension Aliasing #
This mailserver supports the 'plus-extension' mechanism.  Incoming mail for
any valid username with a '+suffix' will be accepted as addressed to the
part of the username preceeding the '+' character.  For example, mail sent to
'username+suffix@[account_domain]' will appear in the inbox of 
'user@[account_domain]'.  Plus-extension aliasing requires no configuration
and is useful in coordination with client filtering rules.

# Address Book Filter Forwarding #
A mechanism exists for automatically adding the sender of a mail message to an
address book filter by forwarding the message to the filterctl address using a
'plus-suffix' to specify the desired book name.  To add the 'From' address of
an email in one of your mail folders to an address book filter, forward the
email to 'filterctl+[book_name]@[account_domain]'.  The filterctl command
processor will first extract the From address from the forwarded message.  The
address book name is then taken from the suffix part of the address.  If the
book does not exist it is created.  Finally, the forwarded message's From
address is added to the address book.  Thereafter, all incoming mail from that
sender will be annotated with the corresponding 'X-Address-Book' header.
`

		usage := "# filterctl subject line commands #\n"
		usage += rule
		for _, cmd := range commands {
			name := cmd.Name
			if cmd.Args != "" {
				name += " " + cmd.Args
			}
			usage += fmt.Sprintf("%s\n%s\n", name, cmd.Detail)
			usage += rule
		}

		sender := viper.GetString("sender")
		messageID, err := DecodedMessageID(viper.GetString("message_id"))
		cobra.CheckErr(err)

		var response APIUsageResponse

		response.User = sender
		response.Request = messageID
		response.Success = true
		response.Message = fmt.Sprintf("%s usage", sender)
		response.Help = strings.Split(help, "\n")
		response.Commands = strings.Split(usage, "\n")

		out, err := json.MarshalIndent(&response, "", "  ")
		cobra.CheckErr(err)
		fmt.Println(string(out))
	},
}

func init() {
	rootCmd.AddCommand(usageCmd)
}
