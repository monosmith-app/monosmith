package sources

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

// captureTimeout bounds a captured command: it only reads or plans.
const captureTimeout = 5 * time.Minute

// CaptureCommand runs a smith command line (smith's own binary, in dir) to its
// end and returns everything it printed. It is how the shell reads the plan of
// `smith add --dry-run` or the choices of `smith choices --json` without
// leaving the interface.
//
// argv is smith's arguments without the program name. It never panics: a
// command that cannot start comes back as Captured with Err set.
func CaptureCommand(dir string, argv []string) Captured {
	ctx, cancel := context.WithTimeout(context.Background(), captureTimeout)
	defer cancel()
	p, err := runner.Start(ctx, runner.Spec{Argv: argv, Dir: dir})
	if err != nil {
		return Captured{Err: err}
	}
	var lines []string
	for {
		batch, exit := p.Next()
		for _, l := range batch {
			lines = append(lines, pane.Strip(l))
		}
		if exit != nil {
			text := strings.TrimRight(strings.Join(lines, "\n"), "\n")
			return Captured{Text: text, OK: exit.Code == 0 && exit.Err == nil && !exit.Cancelled, Err: exit.Err}
		}
	}
}

// OpenPath opens a file with the program the system chooses for it (a
// browser for an HTML page). It does not wait for the program to end.
//
// path is the file to open. It returns the error of starting the opener.
func OpenPath(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
