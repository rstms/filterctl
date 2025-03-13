package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestMessages(t *testing.T) {

	var cases = []struct {
		Name          string
		ExpectSuccess bool
	}{
		{"help", true},
		{"list", true},
		{"reset", true},
		{"delete-ham", true},
		{"delete-all", true},
		{"set", true},
		{"version", true},
		{"fnord", true},
		{"empty", true},
		{"nobody", true},
		{"nosubject", true},
		{"suffix", true},
		{"forwarded", true},
		{"forwarded2", true},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			log.Printf("BEGIN_TEST: %s\n", c.Name)
			input, err := os.ReadFile("testdata/" + c.Name)
			require.Nil(t, err)
			ibuf := bytes.NewBuffer(input)
			obuf := bytes.Buffer{}
			ebuf := bytes.Buffer{}
			cmd := exec.Command("./filterctl", "--config", "testdata/config.yaml")
			cmd.Stdin = ibuf
			cmd.Stdout = &obuf
			cmd.Stderr = &ebuf
			fmt.Printf("Run %s %+v\n", c.Name, cmd)
			runErr := cmd.Run()
			var exitCode int
			if runErr != nil {
				switch e := runErr.(type) {
				case *exec.ExitError:
					exitCode = e.ExitCode()
				default:
					require.Nil(t, runErr)
				}
			} else {
				exitCode = cmd.ProcessState.ExitCode()
			}
			fmt.Printf("Run %s returned: exitCode=%v err=%+v\n", c.Name, exitCode, runErr)
			err = os.WriteFile("testdata/"+c.Name+".out", obuf.Bytes(), 0660)
			require.Nil(t, err)
			err = os.WriteFile("testdata/"+c.Name+".err", ebuf.Bytes(), 0660)
			require.Nil(t, err)
			if c.ExpectSuccess {
				require.Zero(t, exitCode)
			} else {
				require.NotZero(t, exitCode)
			}
		})
	}
}
