package cachesetup

import (
	"bytes"
	"os/exec"
	"strings"
)

// Runner starts a program with text on its standard input and returns what it
// printed. It is the port to the cloud command-line programs: tests pass a
// fake, so no test calls a cloud.
type Runner func(stdin string, argv ...string) (stdout, stderr string, err error)

// ExecRunner is the Runner that starts a real process.
func ExecRunner(stdin string, argv ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = strings.NewReader(stdin)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err = cmd.Run()
	return out.String(), errOut.String(), err
}
