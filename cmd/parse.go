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
	"encoding/json"
	"fmt"
	"github.com/emersion/go-message/mail"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

var RECEIVED_PATTERN = regexp.MustCompile(`^from .* by ([a-z][a-z\.]*) \(OpenSMTPD\) with ESMTPSA .* auth=yes user=([a-z][a-z_-]*) for <filterctl\+*([^@]*)@([^>]+)>.*$`)

var DKIM_DOMAIN_PATTERN = regexp.MustCompile(`d=([a-z\.]*)$`)
var FORWARDED_PATTERN = regexp.MustCompile(`.*----- Forwarded Message -----.*`)

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
Read an email message on stdin, scan headers, and execute the command
specified by the 'Subject' line.  The first word of the subject line
specifies the command, and the following words are passed as command line
arguments.  The subcommand is called with --sender set to the From address.
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

func ParseFile(input io.Reader) error {

	if viper.GetBool("verbose") {
		log.Println("BEGIN-INPUT")
		content, err := ioutil.ReadAll(input)
		cobra.CheckErr(err)
		log.Print(string(content))
		log.Println("END-INPUT")
		input = bytes.NewBuffer(content)
	}

	m, err := mail.CreateReader(input)
	cobra.CheckErr(err)
	printHeaders("message", &m.Header)
	messageID := m.Header.Get("Message-ID")
	if messageID == "" {
		return fmt.Errorf("missing Message-ID header")
	}
	// use the custom request ID header as the messageID if present
	requestID := m.Header.Get("X-Filterctl-Request-Id")
	if requestID == "" {
		requestID = messageID
	}
	requestID = strings.Trim(requestID, "<>")
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
		log.Printf("Message-ID: %s\n", messageID)
		log.Printf("Request-ID: %s\n", requestID)
		log.Println("END-ID")
	}

	if suffix != "" {
		return handleForwardedMessage(m, sender, suffix, requestID)
	}
	return handleCommandMessage(m, sender, requestID)
}

func handleForwardedMessage(m *mail.Reader, sender, suffix, messageID string) error {

	address := parseForwardedBody(m, suffix)

	if address == "" {
		return fmt.Errorf("plus-suffix forwarded from address not found")
	}
	args := []string{"mkaddr", suffix, address}
	log.Printf("handleForwardedMessage: %v", args)
	return ExecuteCommand(sender, messageID, args)
}

func commandHasBodyData(command string) bool {
	switch command {
	case "accounts":
		return true
	case "rescan":
		return true
	case "restore":
		return true
	}
	return false
}

func handleCommandMessage(m *mail.Reader, sender, messageID string) error {
	subject, err := m.Header.Subject()
	cobra.CheckErr(err)
	fields := strings.Split(subject, " ")
	if len(fields) == 0 {
		fields = []string{"help"}
	}

	if len(fields) > 0 {
		command := fields[0]
		if commandHasBodyData(command) {
			filename := parseJSONBody(m, command)
			fields = append(fields, filename)
		}
	}
	return ExecuteCommand(sender, messageID, fields)
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

	//log.Printf("Received: %s\n", received)
	matches := RECEIVED_PATTERN.FindStringSubmatch(received)
	//log.Printf("Matches: %v\n", matches)

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
		if viper.GetBool("insecure_disable_username_check") {
			log.Printf("WARNING: insecure_disable_username_check: %s\n", username)
		} else {
			log.Fatalf("From: invalid user: %s", username)
		}
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
				from := scanForwardedTextBody(p.Body)
				if from != "" {
					return from
				}
			case "text/html":
				from := scanForwardedHTMLBody(p.Body)
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

func parseJSONBody(m *mail.Reader, command string) string {
	if viper.GetBool("verbose") {
		log.Printf("parsing JSON body")
	}
	for {
		p, err := m.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("failure reading mesage body: %v", err)
		}
		if viper.GetBool("verbose") {
			log.Printf("PART: %+v\n", p)
		}
		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			value := h.Get("Content-Type")
			if value != "" {
				contentType, _, _ := strings.Cut(value, ";")
				if contentType != "" && contentType != "text/plain" {
					log.Printf("Warning: unexpected Content-Type: %s\n", contentType)
				}
			}
			return scanJSONBodyToTempFile(p.Body)
		default:
			log.Printf("Warning: unexpected body part header: %v\n", h)

		}
	}
	log.Fatalf("failed parsing JSON body")
	return ""
}

func scanJSONBodyToTempFile(body io.Reader) string {
	data, err := io.ReadAll(body)
	if err != nil {
		log.Fatalf("failed reading message body: %v", err)
	}
	if viper.GetBool("verbose") {
		for i, line := range strings.Split(string(data), "\n") {
			log.Printf("BODY[%n] %s\n", i, line)
		}
	}
	var decoded interface{}
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		log.Fatalf("failed decoding message body as JSON: %v", err)
	}
	formatted, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		log.Fatalf("failed reformatting JSON body data: %v", err)
	}
	tmpFile, err := ioutil.TempFile(os.TempDir(), "filterctl-body-*")
	defer tmpFile.Close()
	if err != nil {
		log.Fatalf("failed creating temp file for JSON body data: %v", err)
	}
	_, err = tmpFile.Write(formatted)
	if err != nil {
		log.Fatalf("failed writing JSON body data to temp file: %v", err)
	}
	filename, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		log.Fatalf("failed converting temp file to absolute pathname: %v", err)
	}
	return filename
}

func scanForwardedTextBody(body io.Reader) string {
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

func scanForwardedHTMLBody(body io.Reader) string {
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
