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
	//"bufio"
	"bufio"
	"bytes"
	"fmt"
	"github.com/emersion/go-message/mail"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"os/user"
	"regexp"
	"strings"
)

// var HEADER_PATTERN = regexp.MustCompile(`^([a-zA-Z _-]*): (.*)$`)
var RECEIVED_PATTERN = regexp.MustCompile(`^from .* by ([a-z][a-z\.]*) \(OpenSMTPD\) with ESMTPSA .* auth=yes user=([a-z][a-z_-]*) for <filterctl(\+[a-zA-Z_-]+){0,1}@([a-z][a-z\.]*)>.*$`)

// var EMAIL_ADDRESS_PATTERN = regexp.MustCompile(`^.* <([a-z][a-z_-]*)@([a-z][a-z\.]*)>$`)
var DKIM_DOMAIN_PATTERN = regexp.MustCompile(`d=([a-z\.]*)$`)
var FORWARDED_PATTERN = regexp.MustCompile(`.*----- Forwarded Message -----.*`)

// var PLUS_SUFFIX_ADDRESS_PATTERN = regexp.MustCompile(`^.* <[a-z][a-z_-]+\+([a-z][a-z_-]+)@[a-z][a-z\.]*>$`)
var MOZ_HEADERS_TABLE_BEGIN_PATTERN = regexp.MustCompile(`class="moz-email-headers-table"`)
var MOZ_HEADERS_TABLE_HEADER_PATTERN = regexp.MustCompile(`<th [^>]*>([a-zA-Z]+): </th>`)
var MOZ_HEADERS_TABLE_ADDRESS_PATTERN = regexp.MustCompile(`.*<a class="moz-txt-link.*" href="mailto:([^"]*)">[^<]*</a>.*`)
var MOZ_HEADERS_TABLE_END_PATTERN = regexp.MustCompile(`</table>`)

var Headers map[string]string
var ReceivedCount int

// parseCmd represents the parse command
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "parse an email message",
	Long: `
Read an email message on stdin, scan headers, and execute the 'class' command.
The subcommand is called with --sender set to the From address, and the
subject line is passed as the rest of the command line.
Header information is used to authorize only locally generated messages.
An email reply is generated containing the output of the command.
If the program is called with no arguments, this subcommand is run by default, 
suitable for inclusion in a .forward file.
`,
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(ParseFile(os.Stdin))
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)
}

func ParseFile(input *os.File) error {

	m, err := mail.CreateReader(input)
	cobra.CheckErr(err)
	printHeaders("message", &m.Header)
	sender, username := checkSender(m.Header)
	checkDKIM(m.Header)
	recipient, suffix := checkRecipient(m.Header)
	checkReceived(m.Header, username, suffix)

	if viper.GetBool("verbose") {
		log.Println("BEGIN-ID")
		log.Printf("Hostname: %s\n", Hostname)
		log.Printf("Domains: %v\n", Domains)
		log.Printf("Username: %s\n", username)
		log.Printf("From: %s\n", sender)
		log.Printf("To: %s\n", recipient)
		log.Printf("Suffix: %s\n", suffix)
		log.Println("END-ID")
	}

	if suffix != "" {
		return handleForwardedMessage(m, sender, suffix)
	}
	return handleCommandMessage(m, sender)
}

func handleForwardedMessage(m *mail.Reader, sender, suffix string) error {

	address := parseForwardedBody(m, suffix)

	if address == "" {
		return fmt.Errorf("plus-suffix forwarded from address not found")
	}
	args := []string{"add", suffix, address}
	log.Printf("handleForwardedMessage: %v", args)
	return ExecuteCommand(sender, args)
}

func handleCommandMessage(m *mail.Reader, sender string) error {
	subject, err := m.Header.Subject()
	cobra.CheckErr(err)
	if subject == "" {
		subject = "help"
	}
	return ExecuteCommand(sender, strings.Split(subject, " "))
}

func printHeaders(name string, header *mail.Header) {
	if viper.GetBool("verbose") {
		log.Printf("BEGIN-HEADERS[%s]\n", name)
		fields := header.Fields()
		for fields.Next() {
			log.Printf("[%s] %s\n", fields.Key(), fields.Value())
		}
		log.Printf("END-HEADERS[%s]\n", name)
	}
}

func checkDKIM(header mail.Header) {

	fields := header.FieldsByKey("Dkim-Signature")
	signature := ""
	for fields.Next() {
		if signature != "" {
			log.Fatal("multiple DKIM signatures detected")
		}
		signature = fields.Value()
	}
	if signature == "" {
		log.Fatal("missing DKIM signature")
	}

	for _, field := range strings.Split(signature, ";") {
		field = strings.TrimSpace(field)
		matches := DKIM_DOMAIN_PATTERN.FindStringSubmatch(field)
		if len(matches) == 2 {
			for _, domain := range Domains {
				if matches[1] == domain {
					return
				}
			}
		}
	}
	log.Fatal("domain not found in DKIM Signature")
}

// verify single received line, matching username and plus-suffix
func checkReceived(header mail.Header, username, suffix string) {

	fields := header.FieldsByKey("Received")
	received := ""
	for fields.Next() {
		if received != "" {
			log.Fatal("multiple Received headers detected")
		}
		received = fields.Value()
	}
	if received == "" {
		log.Fatal("missing Received header")
	}

	matches := RECEIVED_PATTERN.FindStringSubmatch(received)
	/*
		for i, match := range matches {
			log.Printf("match[%d] '%s'\n", i, match)
		}
	*/
	if len(matches) != 5 {
		log.Fatalf("Received: parse failed: %s", received)
	}
	rxHostname := matches[1]
	rxUsername := matches[2]
	rxSuffix := matches[3]
	rxDomain := matches[4]

	rxSuffix = strings.TrimPrefix(rxSuffix, "+")

	if rxHostname != Hostname {
		log.Fatalf("Received: hostname mismatch; expected %s, got %s", Hostname, rxHostname)
	}

	if rxUsername != username {
		log.Fatalf("Received: user mismatch; expected %s, got %s", username, rxUsername)
	}

	if rxSuffix != suffix {
		log.Fatalf("Received: suffix mismatch; expected %s, got %s", suffix, rxSuffix)
	}

	for _, domain := range Domains {
		if rxDomain == domain {
			return
		}
	}
	log.Fatalf("Received: invalid domain: %s", rxDomain)
}

// return fromAddress, username
func checkSender(header mail.Header) (string, string) {
	addrs, err := header.AddressList("From")
	cobra.CheckErr(err)
	if len(addrs) == 0 {
		log.Fatal("missing From: address header")
	}
	if len(addrs) != 1 {
		log.Fatal("From: multiple addresses not allowed")
	}
	address := addrs[0].Address
	parts := strings.Split(address, "@")
	if len(parts) != 2 {
		log.Fatalf("From: unexpected format: %v", addrs)
	}
	username := parts[0]
	domain := parts[1]

	_, err = user.Lookup(username)
	if err != nil {
		log.Fatalf("From: invalid user: %s", username)
	}

	for _, d := range Domains {
		if domain == d {
			return address, username
		}
	}
	log.Fatalf("From: invalid domain: %s", domain)
	return "", ""
}

// return toAddress, plus-suffix
func checkRecipient(header mail.Header) (string, string) {
	addrs, err := header.AddressList("To")
	cobra.CheckErr(err)
	if len(addrs) == 0 {
		log.Fatal("missing To: address header")
	}
	if len(addrs) != 1 {
		log.Fatal("To: multiple addresses not allowed")
	}
	address := addrs[0].Address
	user, _, found := strings.Cut(address, "@")
	if !found {
		log.Fatalf("To: unexpected format: %v", addrs)
	}
	_, suffix, _ := strings.Cut(user, "+")
	return address, suffix
}

func parseForwardedBody(m *mail.Reader, suffix string) string {
	for {
		p, err := m.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("failure parsing forwarded body: %v", err)
		}
		//log.Printf("PART: %+v\n", p)
		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			fromValue := h.Get("From")
			if fromValue != "" {
				addr, err := mail.ParseAddress(fromValue)
				if err != nil {
					log.Fatalf("failed parsing forwarded body part From header: %s", fromValue)
				}
				if viper.GetBool("verbose") {
					log.Printf("Found From address in forwarded body part InlineHeader: %s\n", addr.Address)
				}
				return addr.Address
			}
			value := h.Get("Content-Type")
			contentType, _, _ := strings.Cut(value, ";")
			switch contentType {
			case "text/plain":
				from := scanTextBody(p.Body)
				if from != "" {
					return from
				}
			case "text/html":
				from := scanHTMLBody(p.Body)
				if from != "" {
					return from
				}
			default:
				log.Printf("Warning: unexpected Content-Type: %s\n", contentType)
			}
		default:
			log.Printf("Warning: unexpected forwarded body part header: %v\n", h)

		}
	}
	log.Fatal("failed to locate From address in forwarded body")
	return ""
}

func scanTextBody(body io.Reader) string {
	scanner := bufio.NewScanner(body)
	count := 0
	marker := false
	buf := bytes.Buffer{}
	for scanner.Scan() {
		line := scanner.Text()
		if viper.GetBool("verbose") {
			log.Printf("text[%d]: %s\n", count, line)
		}
		count += 1
		if FORWARDED_PATTERN.MatchString(line) {
			marker = true
			continue
		}
		if marker {
			if strings.TrimSpace(line) == "" {
				break
			}
			_, err := buf.WriteString(line + "\n")
			if err != nil {
				log.Fatalf("failed writing to buffer: %v", err)
			}
		}
	}
	if marker {
		m, err := mail.CreateReader(&buf)
		if err != nil {
			log.Printf("Warning: failed reading forwarded text body as message: %v", err)
			return ""
		}
		//log.Printf("part_message: %+v", m)
		addrs, err := m.Header.AddressList("From")
		if err != nil {
			log.Fatalf("failed reading forwarded text body From: %v", err)
		}
		for _, addr := range addrs {
			if viper.GetBool("verbose") {
				log.Printf("Using From address from reparsed text body headers: %s\n", addr.Address)
			}
			return addr.Address
		}
	}
	return ""
}

func scanHTMLBody(body io.Reader) string {
	scanner := bufio.NewScanner(body)
	count := 0
	marker := false
	table := false
	from := false
	for scanner.Scan() {
		line := scanner.Text()
		if viper.GetBool("verbose") {
			log.Printf("html[%d]: %s\n", count, line)
		}
		count += 1
		if !marker {
			// loop until marker is found
			if FORWARDED_PATTERN.MatchString(line) {
				marker = true
			}
			continue
		}
		if !table {
			// loop until table is found
			if MOZ_HEADERS_TABLE_BEGIN_PATTERN.MatchString(line) {
				//log.Printf("found table start: %s\n", line)
				table = true
			}
			continue
		}
		match := MOZ_HEADERS_TABLE_HEADER_PATTERN.FindStringSubmatch(line)
		if len(match) == 2 {
			//log.Printf("found row: %s\n", match[1])
			if match[1] == "From" {
				from = true
			} else {
				from = false
			}
			continue
		}
		if from {
			// we've passed the from row, look for the address link
			match := MOZ_HEADERS_TABLE_ADDRESS_PATTERN.FindStringSubmatch(line)
			if len(match) == 2 {
				if viper.GetBool("verbose") {
					log.Printf("Parsed From address from html moz-email-headers table: %s\n", match[1])
				}
				return match[1]
			}
		}
		if MOZ_HEADERS_TABLE_END_PATTERN.MatchString(line) {
			//log.Printf("found table end: %s\n", line)
			// we failed to detect a from address, bail out
			return ""
		}
	}
	return ""
}
