package main

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"os"
	"os/exec"
	"testing"
)

func TestMessages(t *testing.T) {

	var cases = []struct {
		Name          string
		ExpectSuccess bool
		OutFile       string
	}{
		{"testdata/help", true, "testdata/help.out"},
		{"testdata/list", true, "testdata/list.out"},
		{"testdata/reset", true, "testdata/reset.out"},
		{"testdata/delete-ham", true, "testdata/delete-ham.out"},
		{"testdata/delete-all", true, "testdata/delete-all.out"},
		{"testdata/set", true, "testdata/set.out"},
		{"testdata/fnord", false, "testdata/fnord.out"},
		{"testdata/empty", false, "testdata/empty.out"},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			input, err := os.ReadFile(c.Name)
			require.Nil(t, err)
			ibuf := bytes.NewBuffer(input)
			obuf := bytes.Buffer{}
			cmd := exec.Command("./filterctl", "--verbose", "--disable-response", "--url", "http://127.0.0.1:2016")
			cmd.Stdin = ibuf
			cmd.Stdout = &obuf
			err = cmd.Run()
			if c.ExpectSuccess {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
			}
			err = os.WriteFile(c.OutFile, obuf.Bytes(), 0660)
			require.Nil(t, err)
		})
	}
}
