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
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var logFile *os.File
var Debug bool
var Verbose bool

const Version = "0.0.10"

var Domains []string
var Hostname string
var Username string
var Sender string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "filterctl",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		file, err := os.OpenFile("filterctl.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		cobra.CheckErr(err)
		logFile = file
		log.SetOutput(logFile)

		err = InitIdentity()
		cobra.CheckErr(err)
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := ParseFile(os.Stdin)
		cobra.CheckErr(err)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if logFile != nil {
			err := logFile.Close()
			cobra.CheckErr(err)
			logFile = nil
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.filterctl.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "disable", "d", false, "disable sending output as mail message")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "enable diagnostic output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".filterctl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".filterctl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func InitIdentity() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return err
	}
	pattern, err := regexp.Compile(`^([a-z][a-z]*)\.([a-z\.]*)\.$`)
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		names, err := net.LookupAddr(addr)
		if err != nil {
			return err
		}
		if len(names) != 1 {
			return fmt.Errorf("unexpected multiple names returned for %s", addr)
		}
		matches := pattern.FindStringSubmatch(names[0])
		if len(matches) != 3 {
			return fmt.Errorf("unexpected domain format: %s", names[0])
		}
		host := matches[1]
		domain := matches[2]
		if host == hostname {
			Hostname = host + "." + domain
		}
		Domains = append(Domains, domain)
	}
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	Username = currentUser.Username
	return nil
}

func ExecuteCommand(cmdline string) error {
	args := strings.Split(cmdline, " ")
	fmt.Printf("sender=%s command=%s args=%s\n", Sender, os.Args[0], cmdline)
	cmd := exec.Command(os.Args[0], args...)
	result, err := run(cmd)
	if err != nil {
		return err
	}
	if Verbose {
		log.Println("BEGIN_RESULT")
		log.Println(string(result))
		log.Println("END_RESULT")
	}
	if Debug {
		return nil
	}
	sendmail := exec.Command("sendmail", Sender)
	sendmail.Stdin = bytes.NewBuffer(result)
	_, err = run(sendmail)
	return err
}

func run(cmd *exec.Cmd) ([]byte, error) {
	exitCode := -1
	var oBuf bytes.Buffer
	var eBuf bytes.Buffer
	cmd.Stdout = &oBuf
	cmd.Stderr = &eBuf
	err := cmd.Run()
	if err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			exitCode = e.ExitCode()
		default:
			return nil, fmt.Errorf("subprocess exec failed: %v", err)
		}
	} else {
		exitCode = cmd.ProcessState.ExitCode()
	}

	if exitCode != 0 {
		return nil, fmt.Errorf("subprocess exited %d\n%s\n", exitCode, eBuf.String())
	}
	if eBuf.Len() > 0 {
		return nil, fmt.Errorf("subprocess emitted stderr\n%s\n", eBuf.String())
	}
	return oBuf.Bytes(), nil
}
