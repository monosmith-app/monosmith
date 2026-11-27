//go:build !windows

package versionhandoff

import "syscall"

// startProgram replaces this process with program, so no smith is left in front
// of it: the terminal, the signals and the exit code are the program's own.
// It returns only when the system could not start the program.
func startProgram(program string, args, env []string) (int, error) {
	err := syscall.Exec(program, append([]string{program}, args...), env)
	return 1, err
}
