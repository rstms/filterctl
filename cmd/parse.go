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
var RECEIVED_PATTERN = regexp.MustCompile(`^from .* by ([a-z][a-z\.]*) \(OpenSMTPD\) with ESMTPSA .* auth=yes user=([a-z][a-z_-]*) for <filterctl@([a-z][a-z\.]*)>.*$`)
var FROM_PATTERN = regexp.MustCompile(`^.* <([a-z][a-z_-]*)@([a-z][a-z\.]*)>$`)
var DKIM_DOMAIN_PATTERN = regexp.MustCompile(`d=([a-z\.]*)$`)

var Headers map[string]string
var LastHeader string
var ReceivedCount int

// parseCmd represents the parse command
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := ParseFile(os.Stdin)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// parseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// parseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ParseFile(input *os.File) error {
	Headers = make(map[string]string)
	ReceivedCount = 0
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		err, done := parseLine(scanner.Text())
		cobra.CheckErr(err)
		if done {
			break
		}
	}

	if viper.GetBool("verbose") {
		log.Println("BEGIN-HEADERS")
		for header, value := range Headers {
			log.Printf("[%s] %s\n", header, value)
		}
		log.Println("END-HEADERS")
	}

	err := checkHeaders(Headers)
	cobra.CheckErr(err)

	if viper.GetBool("verbose") {
		log.Println("BEGIN-ID")
		log.Printf("Hostname: %s\n", Hostname)
		log.Printf("Username: %s\n", Username)
		log.Printf("Domains: %v\n", Domains)
		log.Printf("Sender: %s\n", Sender)
		log.Println("END-ID")
	}

	return ExecuteCommand(Headers["Subject"])
}

func parseLine(line string) (error, bool) {

	isHeader, err := regexp.MatchString(`^[a-zA-Z]`, line)
	if err != nil {
		return err, true
	}
	if !isHeader {
		if LastHeader != "" {
			Headers[LastHeader] = Headers[LastHeader] + " " + strings.TrimSpace(line)
		}
		return nil, false
	}
	matches := HEADER_PATTERN.FindStringSubmatch(line)
	if len(matches) == 3 {
		name := matches[1]
		value := matches[2]
		LastHeader = name
		Headers[name] = value
		if name == "Received" {
			ReceivedCount++
		}
		if name == "Subject" {
			return nil, true
		}
		return nil, false
	}
	return fmt.Errorf("failed to parse: %s", line), true
}

func checkHeaders(headers map[string]string) error {
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
	if len(matches) != 4 {
		return errors.New("Received: parse failed")
	}
	rxHostname := matches[1]
	rxUsername := matches[2]
	rxDomain := matches[3]

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
	matches := FROM_PATTERN.FindStringSubmatch(fromLine)
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
