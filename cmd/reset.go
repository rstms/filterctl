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
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CLASS_PATTERN = regexp.MustCompile(`^\s*([a-zA-Z][a-zA-Z0-9_-]*)=([-0-9\.][0-9\.]*)\s*$`)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset NAME=THRESHOLD [...]",
	Short: "replace rspamd class thresholds",
	Long: `
Replace the set of rspamd class thresholds with a new set provided as
arguments.  Each class name has a threshold value.  The threshold values set
the upper limit for each class.  Any number of classes may be defined.
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		api := initAPI()
		_, err := api.Delete(fmt.Sprintf("/filterctl/classes/%s", viper.GetString("sender")))
		cobra.CheckErr(err)

		for _, arg := range args {
			matches := CLASS_PATTERN.FindStringSubmatch(arg)
			if len(matches) != 3 {
				cobra.CheckErr(fmt.Errorf("failed to parse class specifier '%s'", arg))
			}
			name := matches[1]
			threshold := matches[2]
			_, err := strconv.ParseFloat(threshold, 32)
			if err != nil {
				cobra.CheckErr(fmt.Errorf("invalid threshold value in class specifier '%s' ", arg))
			}
			_, err = api.Put(fmt.Sprintf("/filterctl/classes/%s/%s/%s", viper.GetString("sender"), name, threshold))
			cobra.CheckErr(err)
		}
		response, err := api.Get(fmt.Sprintf("/filterctl/classes/%s", viper.GetString("sender")))
		fmt.Println(response)
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resetCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
