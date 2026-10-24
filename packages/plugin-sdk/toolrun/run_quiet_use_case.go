package toolrun

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// VerboseEnv, when set to anything but "" or "0", makes Quiet stream a tool's
// output as Run does. smith add --verbose sets it for the plugins.
const VerboseEnv = "SMITH_VERBOSE"

// Quiet executes argv in dir like Run, but keeps the tool's output unless it
// fails (or VerboseEnv asks for it): a package manager's warnings about peer
// ranges, engines and deprecated packages are noise to someone who only asked
// for a project, and look like smith's own problems. It says in one line what
// it is doing, so a long install does not look stuck.
func Quiet(dir string, argv []string, extraEnv ...string) error {
	if v := os.Getenv(VerboseEnv); v != "" && v != "0" {
		return Run(dir, argv, extraEnv...)
	}
	cmd, err := command(dir, argv, extraEnv)
	if err != nil {
		return err
	}
	var captured bytes.Buffer
	cmd.Stdout, cmd.Stderr = &captured, &captured
	_, _ = fmt.Fprintf(Log, "Running %s ...\n", strings.Join(argv, " "))
	if err := cmd.Run(); err != nil {
		_, _ = Log.Write(captured.Bytes())
		return fmt.Errorf("%s failed: %w", strings.Join(argv, " "), err)
	}
	return nil
}
