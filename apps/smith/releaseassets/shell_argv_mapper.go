package releaseassets

import (
	"os/exec"
	"runtime"
)

// shellArgv is how a shell runs a command line: sh -c, and on Windows the
// bash that Git for Windows (and every GitHub runner) provides, else cmd.
func shellArgv(command string) []string {
	if runtime.GOOS != "windows" {
		return []string{"sh", "-c", command}
	}
	if _, err := exec.LookPath("bash"); err == nil {
		return []string{"bash", "-c", command}
	}
	return []string{"cmd", "/C", command}
}
