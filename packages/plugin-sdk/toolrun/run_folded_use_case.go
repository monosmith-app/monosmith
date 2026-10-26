package toolrun

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Folded executes argv in dir like Quiet and, when it succeeds, returns its
// warnings folded into one-line notes for the result instead of printing the
// raw blocks: the result stays on top, the details are one flag away
// (VerboseEnv streams them as Run does). A failing tool's whole output goes
// to Log, as it is the reason.
func Folded(dir string, argv []string, extraEnv ...string) ([]string, error) {
	if v := os.Getenv(VerboseEnv); v != "" && v != "0" {
		return nil, Run(dir, argv, extraEnv...)
	}
	cmd, err := command(dir, argv, extraEnv)
	if err != nil {
		return nil, err
	}
	var captured bytes.Buffer
	cmd.Stdout, cmd.Stderr = &captured, &captured
	_, _ = fmt.Fprintf(Log, "Running %s ...\n", strings.Join(argv, " "))
	if err := cmd.Run(); err != nil {
		_, _ = Log.Write(captured.Bytes())
		return nil, fmt.Errorf("%s failed: %w", strings.Join(argv, " "), err)
	}
	return FoldWarnings(captured.String(), "smith install --verbose lists them"), nil
}
