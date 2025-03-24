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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump classes and carddav config",
	Long: `
Return the sender's password, address books and the list of addresses for
each address book.
`,
	Run: func(cmd *cobra.Command, args []string) {
		api := InitAPI()
		var data APIDumpResponse
		sender := viper.GetString("sender")
		path := fmt.Sprintf("/filterctl/dump/%s/", sender)
		_, err := api.Get(path, &data)
		cobra.CheckErr(err)
		_, classes, err := GetClasses(api)
		cobra.CheckErr(err)
		fmt.Printf("classes=%+v\n", classes)
		userData := data.Dump["Users"]
		userData(*map[string]any)
		userData["Classes"] = classes
		fmt.Printf("userData=%+v\n", userData)
		//["Classes"] = classes
		result, err := json.MarshalIndent(&userData, "", "  ")
		fmt.Println(string(result))
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}
