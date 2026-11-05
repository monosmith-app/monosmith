package toolinstall

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// ExecDetector detects a tool by running its program with the version
// arguments. It tries the program, then each alternative, and takes the first
// that runs and exits cleanly. A tool with one wanted version is looked for
// first in the folder where smith installs its own copies for the workspace,
// since the lint task and the doctor look there first too.
type ExecDetector struct{}

// Detect implements Detector.
func (ExecDetector) Detect(ctx context.Context, tool Need) Probe {
	args := tool.VersionArgs
	if len(args) == 0 {
		args = []string{"--version"}
	}
	programs := append([]string{tool.Program}, tool.Alternatives...)
	if tool.Exact && len(tool.Dirs) > 0 && toolrun.ManagedBinExists(tool.Dirs[0]) {
		managed := filepath.Join(toolrun.ManagedBinDir(tool.Dirs[0]), tool.Program)
		programs = append([]string{managed}, programs...)
	}
	for _, program := range programs {
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
