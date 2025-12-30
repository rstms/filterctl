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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var logFile *os.File

const Version = "1.3.16"

var Hostname string
var Username string
var Domains []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "filterctl",
	Short: "mail command processor for rspam class filter",
	Long: `

filterctl is a mail-based command processor for user management of spam class filter settings.

The program is executed as the filterctl user from the /home/filterctl/.forward file:

    |/usr/local/bin/filterctl

It interprets the subject line as a command, executes it, and sends the output as the body
of a new email message sent back to the sender address.

It relies on the mailserver configuration to guarantee that only local mail originating from
a local account via an authorized secure SMTPS session is accepted for the the filterctl user.
In this way it relies on the security of the mailserver's auth mechanism to control access to
the commands.  The sender address is verified as an authorized local user.

Command actions issue API requests to filterctld running at http://localhost:2016/filterctl

`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		err := InitIdentity()
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

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.filterctl.yaml)")

	rootCmd.PersistentFlags().StringP("log-file", "l", "/var/log/filterctl", "log filename")
	viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file"))

	rootCmd.PersistentFlags().BoolP("disable-exec", "d", false, "disable command execution")
	viper.BindPFlag("disable_exec", rootCmd.PersistentFlags().Lookup("disable-exec"))

	rootCmd.PersistentFlags().BoolP("disable-response", "D", false, "disable sending output as mail message")
	viper.BindPFlag("disable_response", rootCmd.PersistentFlags().Lookup("disable-response"))

	rootCmd.PersistentFlags().Bool("insecure-disable-username-check", false, "disable failure if username not valid")
	viper.BindPFlag("insecure_disable_username_check", rootCmd.PersistentFlags().Lookup("insecure-disable-username-check"))

	rootCmd.PersistentFlags().BoolP("version", "", false, "output program name and version")
	viper.BindPFlag("version", rootCmd.PersistentFlags().Lookup("version"))

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable diagnostic output")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.PersistentFlags().String("cert", filepath.Join(home, "/ssl/filterctl.pem"), "client certificate PEM file")
	viper.BindPFlag("cert", rootCmd.PersistentFlags().Lookup("cert"))

	rootCmd.PersistentFlags().String("key", filepath.Join(home, "/ssl/filterctl.key"), "client certificate key file")
	viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key"))

	rootCmd.PersistentFlags().String("ca", "/etc/ssl/keymaster.pem", "certificate authority file")
	viper.BindPFlag("ca", rootCmd.PersistentFlags().Lookup("ca"))

	rootCmd.PersistentFlags().String("server-url", "http://localhost:2016", "server url")
	viper.BindPFlag("server_url", rootCmd.PersistentFlags().Lookup("server-url"))

	rootCmd.PersistentFlags().String("sender", "", "from address")
	viper.BindPFlag("sender", rootCmd.PersistentFlags().Lookup("sender"))

	rootCmd.PersistentFlags().String("message-id", "", "base64-encoded message ID")
	viper.BindPFlag("message_id", rootCmd.PersistentFlags().Lookup("message-id"))

	rootCmd.PersistentFlags().Bool("no-remove", false, "disable deletion of input file")
	viper.BindPFlag("no_remove", rootCmd.PersistentFlags().Lookup("no-remove"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	versionFlag, err := rootCmd.PersistentFlags().GetBool("version")
	cobra.CheckErr(err)
	if versionFlag {
		PrintVersion()
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		config, err := os.UserConfigDir()
		cobra.CheckErr(err)
		userConfigPath := filepath.Join(config, "filterctl")

		// Search config in home directory with name ".filterctl" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(userConfigPath)
		//viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("filterctl")
	}

	viper.SetEnvPrefix("filterctl")
	viper.AutomaticEnv() // read in environment variables that match

	viper.SetDefault("message_id", EncodedMessageID("filter_control_message"))

	// If a config file is found, read it in.
	err = viper.ReadInConfig()
	cobra.CheckErr(err)

	filename := viper.GetString("log_file")
	if filename == "stderr" || filename == "-" {
		log.SetOutput(os.Stderr)
	} else {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		cobra.CheckErr(err)
		logFile = file
		log.SetOutput(logFile)
	}
	log.SetPrefix(fmt.Sprintf("[%d] ", os.Getpid()))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)

	file := viper.ConfigFileUsed()
	if file != "" && viper.GetBool("verbose") {
		log.Printf("Configured from file: %v\n", file)
	}
}

func InitIdentity() error {
	verbose := viper.GetBool("verbose")
	Hostname = viper.GetString("hostname")
	Domains = viper.GetStringSlice("domains")
	if verbose {
		log.Printf("viper hostname: %s\n", Hostname)
		log.Printf("viper domains: %v\n", Domains)
	}
	if Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		if verbose {
			log.Printf("autoconfig hostname=%v\n", hostname)
		}
		addrs, err := net.LookupHost(hostname)
		if err != nil {
			return err
		}
		if verbose {
			log.Printf("autoconfig addrs=%v\n", addrs)
		}
		for _, addr := range addrs {
			names, err := net.LookupAddr(addr)
			if err != nil {
				return err
			}
			if verbose {
				log.Printf("autoconfig names=%v\n", names)
			}
			for _, name := range names {
				parts := strings.Split(name, ".")
				log.Printf("autoconfig name=%v parts=%v\n", name, parts)
				if len(parts) < 4 || parts[len(parts)-1] != "" {
					return errors.New(fmt.Sprintf("unexpected domain format: for %s: %v", addr, parts))
				}
				host := parts[0]
				domain := strings.Join(parts[1:len(parts)-1], ".")
				if verbose {
					log.Printf("addr=%s hostname=%s host=%s domain=%s\n", addr, hostname, host, domain)
				}
				if Hostname == "" {
					Hostname = host + "." + domain
				}
				Domains = append(Domains, domain)
			}
		}
	}
	if verbose {
		log.Printf("Hostname: %s\n", Hostname)
		log.Printf("Domains: %v\n", Domains)
	}

	if Hostname == "" {
		return errors.New("failed to set Hostname")
	}
	if len(Domains) == 0 {
		return errors.New("failed to set Domain")
	}
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	Username = currentUser.Username
	return nil
}

func LogLines(label string, buf []byte) {
	if len(buf) > 0 {
		log.Printf("BEGIN_%s\n", label)
		for i, line := range strings.Split(string(buf), "\n") {
			log.Printf("%03d: %s\n", i, line)
		}
		log.Printf("END_%s\n", label)
	}
}

func EncodedMessageID(messageID string) string {
	return base64.StdEncoding.EncodeToString([]byte(messageID))
}

func DecodedMessageID(encoded string) (string, error) {
	if encoded == "" {
		return "", fmt.Errorf("encoded message_id is null")
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("message_id decode failed: %v", err)
	}
	return string(decoded), nil
}

func ExecuteCommand(sender, messageID string, args []string) error {
	verbose := viper.GetBool("verbose")
	if verbose {
		log.Printf("ExecuteCommand: sender=%s messageID=%s command=%s args=%v\n", sender, messageID, os.Args[0], args)
	}
	if viper.GetBool("disable_exec") {
		return nil
	}

	if args[0] == "help" {
		args[0] = "usage"
	}
	viper.Set("sender", sender)
	viper.Set("message_id", EncodedMessageID(messageID))
	cmd := exec.Command(os.Args[0], args...)

	cmd.Env = []string{}
	for _, key := range []string{"HOME", "PATH", "TERM"} {
		value := fmt.Sprintf("%s=%s", key, os.Getenv(key))
		cmd.Env = append(cmd.Env, value)
	}
	for key, value := range viper.AllSettings() {
		value := fmt.Sprintf("FILTERCTL_%s=%v", strings.ToUpper(key), value)
		cmd.Env = append(cmd.Env, value)
	}

	if verbose {
		log.Println("BEGIN_SUBPROCESS_ENV")
		for _, value := range cmd.Environ() {
			log.Println(value)
		}
		log.Println("END_SUBPROCESS_ENV")
	}

	exitCode, stdout, stderr, err := run(cmd)

	if verbose {
		log.Printf("subprocess exited %d\n", exitCode)
		LogLines("SUBPROCESS_STDOUT", stdout)
		LogLines("SUBPROCESS_STDERR", stderr)
	}

	if err != nil || exitCode != 0 {
		fail := map[string]any{
			"Success": false,
			"Request": messageID,
			"Message": fmt.Sprintf("%s internal failure", sender),
			"Help":    "Send 'help' in Subject line for valid commands",
		}
		if verbose {
			detail := map[string]any{}
			if err != nil {
				detail["err"] = fmt.Sprintf("%v", err)
			}
			detail["exit"] = exitCode
			ostr := strings.TrimSpace(string(stdout))
			if len(ostr) > 0 {
				detail["stdout"] = strings.Split(ostr, "\n")
			}
			estr := strings.TrimSpace(string(stderr))
			if len(estr) > 0 {
				detail["stderr"] = strings.Split(estr, "\n")
			}
		}
		result, err := json.MarshalIndent(&fail, "", "  ")
		if err != nil {
			return err
		}
		stdout = result
	}

	// generate RFC2822 email message
	responseSubject := fmt.Sprintf("filterctl response %s", viper.GetString("message-id"))
	message, err := formatEmailMessage(messageID, responseSubject, sender, "filterctl@"+Domains[0], stdout)
	if err != nil {
		return err
	}

	if verbose {
		LogLines("RESPONSE", message)
	}

	if viper.GetBool("disable_response") {
		fmt.Println(string(message))
		return nil
	}

	sendmail := exec.Command("sendmail", sender)
	sendmail.Stdin = bytes.NewBuffer(message)
	exitCode, stdout, stderr, err = run(sendmail)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		log.Printf("sendmail exited %d\n", exitCode)
		LogLines("SENDMAIL_STDOUT", stdout)
		LogLines("SENDMAIL_STDERR", stderr)
	}
	return nil
}

func run(cmd *exec.Cmd) (int, []byte, []byte, error) {
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
			return -1, nil, nil, err
		}
	} else {
		exitCode = cmd.ProcessState.ExitCode()
	}
	return exitCode, oBuf.Bytes(), eBuf.Bytes(), nil
}

func NewFilterctlClient() *APIClient {
	url := viper.GetString("server_url")
	api, err := NewAPIClient(url)
	cobra.CheckErr(err)
	if viper.GetString("sender") == "" {
		cobra.CheckErr(errors.New("missing sender"))
	}
	return api
}

func PrintVersion() {
	fmt.Printf("filterctl version %s\n", Version)
	os.Exit(0)
}
