//go:build windows

package shellenv

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// parentProgramName is the executable name of process pid, read from a
// snapshot of the running processes; "" when the process is gone.
func parentProgramName(pid int) string {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(snap)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for err = windows.Process32First(snap, &entry); err == nil; err = windows.Process32Next(snap, &entry) {
		if int(entry.ProcessID) == pid {
			return windows.UTF16ToString(entry.ExeFile[:])
		}
	}
	return ""
}
