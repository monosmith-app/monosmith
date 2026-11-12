package toolinstall

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// ExecDetector detects a tool by running its program with the version
// arguments. It tries the program, then each alternative, and takes the first
// that runs and exits cleanly. A tool with one wanted version is looked for
// first in the folder where smith installs its own copies for the workspace,
// since the lint task and the doctor look there first too. On Windows, a tool
// that no program on PATH answers for is looked for in the folders where
// winget unpacks a portable package (FindWingetFolder): the probe then names
// that folder, so smith can offer to use it.
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
	if runtime.GOOS != "windows" {
		return Probe{}
	}
	dir := FindWingetFolder(os.Getenv("LOCALAPPDATA"), tool)
	if dir == "" {
		return Probe{}
	}
	probe := Probe{Folder: dir}
	for _, program := range append([]string{tool.Program}, tool.Alternatives...) {
		out, err := exec.CommandContext(ctx, filepath.Join(dir, program+".exe"), args...).CombinedOutput()
		if err == nil {
			probe.Version = FindVersion(string(out))
			break
		}
	}
	return probe
}
