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
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set CLASS=THRESHOLD",
	Short: "set a single class name and threshold",
	Long: `
Add or update a single class name and threshold value.
CLASS is an identifier string.
THRESHOLD is a floating point number.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filterctl := NewFilterctlClient()
		var response APIResponse
		class := args[0]
		matches := CLASS_PATTERN.FindStringSubmatch(class)

		if len(matches) != 3 {
			cobra.CheckErr(fmt.Errorf("failed to parse class specifier '%s'", class))
		}
		name := matches[1]
		threshold := matches[2]
		_, err := strconv.ParseFloat(threshold, 32)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("invalid threshold value in class specifier '%s' ", class))
		}

		text, err := filterctl.Put(fmt.Sprintf("/filterctl/classes/%s/%s/%s/", viper.GetString("sender"), name, threshold), &response)
		cobra.CheckErr(err)
		fmt.Println(text)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
