package taskrun

import (
	"os/exec"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var jobs sync.Map // *exec.Cmd -> windows.Handle

// prepare starts the child in a new process group so Ctrl+C in the terminal
// reaches smith first, and smith decides how to stop the tree.
func prepare(cmd *exec.Cmd) {
	cmd.SysProcAttr = &windows.SysProcAttr{CreationFlags: windows.CREATE_NEW_PROCESS_GROUP}
}

// track puts the started child in a Job Object that kills every process in it
// when the job is closed, so no grandchild (node, tsc, python) outlives smith.
func track(cmd *exec.Cmd) (release func(), err error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return func() {}, err
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		_ = windows.CloseHandle(job)
		return func() {}, err
	}
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		_ = windows.CloseHandle(job)
		return func() {}, err
	}
	defer func() { _ = windows.CloseHandle(process) }()
	if err := windows.AssignProcessToJobObject(job, process); err != nil {
		_ = windows.CloseHandle(job)
		return func() {}, err
	}
	jobs.Store(cmd, job)
	return func() {
		if h, ok := jobs.LoadAndDelete(cmd); ok {
			_ = windows.CloseHandle(h.(windows.Handle))
		}
	}, nil
}

// forceKillTree ends the job again: Windows has no gentle first step.
func forceKillTree(cmd *exec.Cmd) { killTree(cmd) }

// killTree terminates every process in the child's job.
func killTree(cmd *exec.Cmd) {
	if h, ok := jobs.Load(cmd); ok {
		_ = windows.TerminateJobObject(h.(windows.Handle), 1)
		return
	}
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
