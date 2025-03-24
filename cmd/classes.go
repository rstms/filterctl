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
	"github.com/spf13/viper"
)

var classesCmd = &cobra.Command{
	Use:   "classes",
	Short: "list rspamd classes",
	Long: `
Return the complete set of rspamd class names and threshold values for the
sender address.
`,
	Run: func(cmd *cobra.Command, args []string) {
		api := InitAPI()
		classes, _, err := GetClasses(api)
		cobra.CheckErr(err)
		fmt.Println(classes)
	},
}

func init() {
	rootCmd.AddCommand(classesCmd)
}

func GetClasses(api *APIClient) (string, *APIClassesResponse, error) {
	var data APIClassesResponse
	path := fmt.Sprintf("/filterctl/classes/%s/", viper.GetString("sender"))
	response, err := api.Get(path, &data)
	if err != nil {
		return "", nil, err
	}
	return response, &data, nil
}
