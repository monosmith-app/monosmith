package toolpin

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	tableHeader = regexp.MustCompile(`^\s*\[\s*([^\]]*?)\s*\]\s*(#.*)?$`)
	pinLine     = regexp.MustCompile(`^\s*"?([A-Za-z0-9_.-]+)"?\s*=\s*(?:"([^"]*)"|'([^']*)')\s*(#.*)?$`)
)

// smithPin reads the [toolchains] pin of a label from smith.toml. It is empty
// when the file, the table or the label is missing.
func smithPin(workspaceRoot, label string) string {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "smith.toml"))
	if err != nil {
		return ""
	}
	in := false
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, "\r")
		if m := tableHeader.FindStringSubmatch(line); m != nil {
			in = m[1] == "toolchains"
			continue
		}
		if !in {
			continue
		}
		if m := pinLine.FindStringSubmatch(line); m != nil && m[1] == label {
			return strings.TrimSpace(m[2] + m[3])
		}
	}
	return ""
}
