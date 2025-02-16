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
	"bufio"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/user"
	"regexp"
	"strings"
)

var HEADER_PATTERN = regexp.MustCompile(`^([a-zA-Z _-]*): (.*)$`)
var RECEIVED_PATTERN = regexp.MustCompile(`^from .* by ([a-z][a-z\.]*) \(OpenSMTPD\) with ESMTPSA .* auth=yes user=([a-z][a-z_-]*) for <filterctl(\+[a-zA-Z_-]+){0,1}@([a-z][a-z\.]*)>.*$`)
var EMAIL_ADDRESS_PATTERN = regexp.MustCompile(`^.* <([a-z][a-z_-]*)@([a-z][a-z\.]*)>$`)
var DKIM_DOMAIN_PATTERN = regexp.MustCompile(`d=([a-z\.]*)$`)
var FORWARDED_PATTERN = regexp.MustCompile(`.*----- Forwarded Message -----.*`)

/*
 */

var Headers map[string]string
var ForwardHeaders map[string]string
var LastHeader string
var ReceivedCount int
var PlusSuffix string

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
		err := ParseFile(os.Stdin)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)
}

func ParseFile(input *os.File) error {

	if viper.GetBool("verbose") {
		log.Println("BEGIN-MESSAGE")
	}
	var forwardTo, forwardFrom string
	Headers = make(map[string]string)
	ForwardHeaders = make(map[string]string)
	ReceivedCount = 0
	lineCount := 0
	scanner := bufio.NewScanner(input)
	inHeader := true
	scanComplete := false
	log.Println("BEGIN parseHeaders")
	for scanner.Scan() {
		line := scanner.Text()
		if viper.GetBool("verbose") {
			log.Printf("%03d: %s\n", lineCount, string(line))
		}
		lineCount += 1

		if inHeader {
			done, err := parseHeaderLine(line, false)
			cobra.CheckErr(err)
			if done {
				inHeader = false
				log.Println("END parseHeaders")
				err := checkHeaders()
				cobra.CheckErr(err)
				log.Printf("Headers[To]: %s\n", Headers["To"])
				log.Printf("Headers[X-Plus-Suffix]: %s\n", Headers["X-Plus-Suffix"])
				if Headers["X-Plus-Suffix"] != "" {
					log.Println("BEGIN parseForwardedBody")
				} else {
					scanComplete = true
				}
			}
		} else {
			done, err := parseForwardedBodyLine(line)
			cobra.CheckErr(err)
			if done {
				log.Println("END parseForwardedBody")
				scanComplete = true
				forwardTo, forwardFrom, err = checkForwardHeaders()
				cobra.CheckErr(err)
			}
		}

		if scanComplete {
			break
		}

	}

	if viper.GetBool("verbose") {
		log.Println("END-MESSAGE")
	}

	printHeaders("message", &Headers)
	printHeaders("forward", &ForwardHeaders)

	if viper.GetBool("verbose") {
		log.Println("BEGIN-ID")
		log.Printf("Hostname: %s\n", Hostname)
		log.Printf("Username: %s\n", Username)
		log.Printf("Domains: %v\n", Domains)
		log.Printf("Sender: %s\n", Sender)
		log.Printf("To: %s\n", Headers["To"])
		log.Printf("Suffix: %s\n", Headers["X-Plus-Suffix"])
		log.Println("END-ID")
	}

	var args []string
	if Headers["X-Plus-Suffix"] != "" {
		// plus suffix was used
		if forwardTo == "" {
			return fmt.Errorf("plus-suffix is only valid with forwarded messages")
		}
		book := strings.TrimPrefix(Headers["X-Plus-Suffix"], "+")
		if book == "" {
			return fmt.Errorf("null plus-suffix address book")
		}
		address := forwardFrom
		if address == "" {
			return fmt.Errorf("plus-suffix forwarded from address not found")
		}
		args = []string{"add", book, address}
	} else {
		// this is a command message
		subject := strings.TrimSpace(Headers["Subject"])
		if subject == "" {
			subject = "help"
		}
		args = strings.Split(subject, " ")
	}
	return ExecuteCommand(args)
}

func printHeaders(name string, headers *map[string]string) {
	if viper.GetBool("verbose") {
		log.Printf("BEGIN-HEADERS[%s]", name)
		for header, value := range *headers {
			log.Printf("[%s] %s\n", header, value)
		}
		log.Printf("END-HEADERS[%s]", name)
	}
}

func parseHeaderLine(line string, forward bool) (bool, error) {

	headers := &Headers
	if forward {
		// in forwarded body, save to ForwardHeaders
		headers = &ForwardHeaders
	}

	// blank line terminates headers
	if len(strings.TrimSpace(line)) == 0 {
		return true, nil
	}

	isHeader, err := regexp.MatchString(`^[a-zA-Z0-9]`, line)
	if err != nil {
		return true, err
	}
	if !isHeader {
		if LastHeader != "" {
			(*headers)[LastHeader] = (*headers)[LastHeader] + " " + strings.TrimSpace(line)
		}
		return false, nil
	}
	matches := HEADER_PATTERN.FindStringSubmatch(line)
	if len(matches) == 3 {
		name := matches[1]
		value := matches[2]
		LastHeader = name
		(*headers)[name] = value
		if !forward {
			// if not in forwarded body headers, count Received lines
			if name == "Received" {
				ReceivedCount++
			}
		}
		return false, nil
	}
	return true, fmt.Errorf("failed to parse: %s", line)
}

func parseForwardedBodyLine(line string) (bool, error) {
	if FORWARDED_PATTERN.MatchString(line) {
		// skip until we find the forwarded marker
		ForwardHeaders["X-Forwarded-Marker"] = strings.TrimSpace(line)
		return false, nil
	}

	if ForwardHeaders["X-Forwarded-Marker"] != "" {
		// parse headers following the marker
		return parseHeaderLine(line, true)
	}
	return false, nil
}

func checkForwardHeaders() (string, string, error) {
	var to, from string
	line, ok := ForwardHeaders["To"]
	if ok {
		matches := EMAIL_ADDRESS_PATTERN.FindStringSubmatch(line)
		if len(matches) > 2 {
			to = fmt.Sprintf("%s@%s", matches[1], matches[2])
		}
	}
	line, ok = ForwardHeaders["From"]
	if ok {
		matches := EMAIL_ADDRESS_PATTERN.FindStringSubmatch(line)
		if len(matches) > 2 {
			from = fmt.Sprintf("%s@%s", matches[1], matches[2])
		}

	}

	if to == "" {
		return "", "", fmt.Errorf("missing forwarded To header")
	}

	if from == "" {
		return "", "", fmt.Errorf("missing forwarded From header")
	}

	return to, from, nil
}

func checkHeaders() error {
	err := checkDKIM()
	if err != nil {
		return err
	}
	err = checkSender()
	if err != nil {
		return err
	}
	err = checkReceived()
	if err != nil {
		return err
	}
	return nil
}

func checkDKIM() error {

	dkim, ok := Headers["DKIM-Signature"]
	if !ok {
		return errors.New("missing DKIM signature")
	}
	for _, field := range strings.Split(dkim, ";") {
		field = strings.TrimSpace(field)
		matches := DKIM_DOMAIN_PATTERN.FindStringSubmatch(field)
		if len(matches) == 2 {
			for _, domain := range Domains {
				if matches[1] == domain {
					return nil
				}
			}
		}
	}
	return errors.New("domain not found in DKIM Signature")
}

func checkReceived() error {
	if ReceivedCount != 1 {
		return fmt.Errorf("Received: bad count; expected 1,  got %d", ReceivedCount)
	}
	received, ok := Headers["Received"]
	if !ok {
		return errors.New("Received: missing header")
	}
	matches := RECEIVED_PATTERN.FindStringSubmatch(received)
	for i, match := range matches {
		log.Printf("match[%d] '%s'\n", i, match)
	}
	if len(matches) != 5 {
		return errors.New("Received: parse failed")
	}
	rxHostname := matches[1]
	rxUsername := matches[2]
	Headers["X-Plus-Suffix"] = matches[3]
	rxDomain := matches[4]

	if rxHostname != Hostname {
		return fmt.Errorf("Received: hostname mismatch; expected %s, got %s", Hostname, rxHostname)
	}

	if rxUsername != Username {
		return fmt.Errorf("Received: user mismatch; expected %s, got %s", Username, rxUsername)
	}
	for _, domain := range Domains {
		if rxDomain == domain {
			return nil
		}
	}
	return fmt.Errorf("Received: invalid domain: %s", rxDomain)
}

func checkSender() error {
	fromLine, ok := Headers["From"]
	if !ok {
		return errors.New("From: missing header")
	}
	matches := EMAIL_ADDRESS_PATTERN.FindStringSubmatch(fromLine)
	if len(matches) != 3 {
		return errors.New("From: parse failed")
	}
	fromUser := matches[1]
	fromDomain := matches[2]

	_, err := user.Lookup(fromUser)
	if err != nil {
		return fmt.Errorf("From: invalid user: %s", fromUser)
	}

	Username = fromUser

	for _, domain := range Domains {
		if domain == fromDomain {
			Sender = fromUser + "@" + fromDomain
			return nil
		}
	}
	return fmt.Errorf("From: invalid domain: %s", fromDomain)
}
