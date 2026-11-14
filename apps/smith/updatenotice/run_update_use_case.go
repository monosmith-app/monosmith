package updatenotice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

// ErrNotRunnable is returned when smith cannot tell how it was installed, so
// there is no command to run.
var ErrNotRunnable = errors.New("smith cannot tell how it was installed")

// Run installs the new smith by running the command that matches how smith was
// installed. On Windows the binary of the running smith cannot be replaced, so
// the command runs in a new console window that first waits for this process to
// exit: the caller must exit right after Run returns detached = true. Other
// systems run the command here, with the terminal of the caller.
//
// Parameters: command is Update.Command; goos is runtime.GOOS; pid is this
// process id; out and errOut are where the command writes when it runs here.
// Returns: detached, true when the command continues in a new window after this
// process exits; an error when the command is not runnable or fails.
func Run(ctx context.Context, command, goos string, pid int, out, errOut io.Writer) (detached bool, err error) {
	if command == "" || isReleasePage(command) {
		return false, ErrNotRunnable
	}
	if goos == "windows" {
		script := "Wait-Process -Id " + strconv.Itoa(pid) + " -ErrorAction SilentlyContinue; " + command +
			"; Write-Host ''; Write-Host 'smith update finished. Press Enter to close this window.'; [void](Read-Host)"
		if err := exec.Command("cmd", "/c", "start", "smith update", "powershell", "-NoProfile", "-Command", script).Start(); err != nil {
			return false, fmt.Errorf("could not open the update window: %w", err)
		}
		return true, nil
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = out, errOut, os.Stdin
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("%s failed: %w", command, err)
	}
	return false, nil
}
