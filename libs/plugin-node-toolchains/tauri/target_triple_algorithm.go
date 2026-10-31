package tauri

import (
	"os/exec"
	"strings"
)

// HostTriple is the target triple Tauri builds for on this machine, from
// `rustc -vV`; a test replaces it.
var HostTriple = func() (string, error) {
	out, err := exec.Command("rustc", "-vV").Output()
	if err != nil {
		return "", err
	}
	return hostTripleOf(string(out)), nil
}

// hostTripleOf reads the host triple from the output of `rustc -vV`.
func hostTripleOf(rustcVV string) string {
	for _, line := range strings.Split(rustcVV, "\n") {
		if t, ok := strings.CutPrefix(strings.TrimSpace(line), "host: "); ok {
			return t
		}
	}
	return ""
}

// exeSuffix is ".exe" for a Windows target triple.
func exeSuffix(triple string) string {
	if strings.Contains(triple, "windows") {
		return ".exe"
	}
	return ""
}
