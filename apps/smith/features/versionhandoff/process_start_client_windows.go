package versionhandoff

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
)

// startProgram runs program on this terminal and waits for it: Windows cannot
// replace a process. Ctrl+C reaches both processes, so this one ignores it while
// it waits and gives back the exit code of the program.
func startProgram(program string, args, env []string) (int, error) {
	cmd := exec.Command(program, args...)
	cmd.Env = env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	defer signal.Stop(interrupts)
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), nil
	}
	if err != nil {
		return 1, err
	}
	return 0, nil
}
