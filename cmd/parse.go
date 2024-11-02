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
    "os"
    "fmt"
    "strings"
    "log"
    "bufio"
    "regexp"
	"github.com/spf13/cobra"
)

var HEADER_PATTERN = regexp.MustCompile(`^([a-zA-Z _-]*): (.*)$`)
var Headers map[string]string
var LastHeader string

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
	    scanner := bufio.NewScanner(input)
	    for scanner.Scan() {
		err, done := parseLine(scanner.Text())
		cobra.CheckErr(err)
		if done {
		    break;
		}
	    }
	    log.Println("BEGIN")
	    for header, value := range Headers {
		log.Printf("[%s] %s\n", header, value)
	    }
	    log.Println("END")
	    return nil
}

func parseLine(line string) (error, bool) {
    
    isHeader, err := regexp.MatchString(`^[a-zA-Z]`, line)
    if err != nil {
	return err, true
    }
    fmt.Printf("isHeader=%v line=%s\n", isHeader, line)
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
	if name == "Subject" {
	    return nil, true
	}
	return nil, false
    }
    return fmt.Errorf("failed to parse: %s", line), true
}
