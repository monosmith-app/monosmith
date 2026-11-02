package toolinstall

import (
	"context"
	"os/exec"
)

// ExecDetector detects a tool by running its program with the version
// arguments. It tries the program, then each alternative, and takes the first
// that runs and exits cleanly.
type ExecDetector struct{}

// Detect implements Detector.
func (ExecDetector) Detect(ctx context.Context, tool Need) Probe {
	args := tool.VersionArgs
	if len(args) == 0 {
		args = []string{"--version"}
	}
	for _, program := range append([]string{tool.Program}, tool.Alternatives...) {
		if _, err := exec.LookPath(program); err != nil {
			continue
		}
		out, err := exec.CommandContext(ctx, program, args...).CombinedOutput()
		if err != nil {
			continue // a Windows Store stub or a broken shim: try the next program
		}
		return Probe{Program: program, Version: FindVersion(string(out))}
	}
	return Probe{}
}
