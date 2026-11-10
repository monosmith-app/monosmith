//go:build !windows

package taskrun

import (
	"os/exec"
	"syscall"
)

// prepare puts the child in its own process group so the whole tree can be
// signalled.
func prepare(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// track has nothing to do on Unix: the process group is the tree.
func track(*exec.Cmd) (release func(), err error) { return func() {}, nil }

// killTree sends SIGTERM to the whole process group so every descendant can
// shut down. forceKillTree follows it when the group does not exit in time.
func killTree(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
}

// forceKillTree sends SIGKILL to the whole process group.
func forceKillTree(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
