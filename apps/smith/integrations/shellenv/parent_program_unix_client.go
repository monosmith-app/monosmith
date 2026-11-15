//go:build !windows

package shellenv

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// parentProgramName is the name of the program of process pid. It reads
// /proc where that exists and asks ps elsewhere.
func parentProgramName(pid int) string {
	if raw, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/comm"); err == nil {
		return strings.TrimSpace(string(raw))
	}
	out, err := exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
