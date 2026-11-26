//go:build !windows

package taskrun

import (
	"os/exec"
	"syscall"
)

// prepare puts the child in its own process group so the whole tree can be
// signalled. A child on the terminal stays in smith's group, the terminal's
// foreground one: in a group of its own it could not read the terminal and
// Ctrl+C would not reach it.
func prepare(cmd *exec.Cmd, terminal bool) {
	if terminal {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// target is the pid killTree signals: the child's process group (negative)
// when it has one, else the child alone.
func target(cmd *exec.Cmd) int {
	if cmd.SysProcAttr != nil && cmd.SysProcAttr.Setpgid {
		return -cmd.Process.Pid
	}
	return cmd.Process.Pid
}

// track has nothing to do on Unix: the process group is the tree.
func track(*exec.Cmd) (release func(), err error) { return func() {}, nil }

// killTree sends SIGTERM to the whole process group so every descendant can
// shut down. forceKillTree follows it when the group does not exit in time.
func killTree(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(target(cmd), syscall.SIGTERM)
	}
}

// forceKillTree sends SIGKILL to the whole process group.
func forceKillTree(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(target(cmd), syscall.SIGKILL)
	}
}
