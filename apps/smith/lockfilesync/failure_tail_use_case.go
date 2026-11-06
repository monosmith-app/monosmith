package lockfilesync

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	outputTailLines  = 30
	logFileTailLines = 40
)

var namedLog = regexp.MustCompile(`(?m)complete log of this run can be found in:\s*(\S.*?)\s*$`)

// lastLines is the last n lines of text.
func lastLines(text string, n int) string {
	lines := strings.Split(strings.TrimRight(strings.ReplaceAll(text, "\r\n", "\n"), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// failureTail is what to show when npm fails: the end of npm's own log file
// when its output names one (npm prints "A complete log of this run can be
// found in: <path>"), then the end of its output.
func failureTail(output string) string {
	var b strings.Builder
	if all := namedLog.FindAllStringSubmatch(output, -1); len(all) > 0 {
		path := all[len(all)-1][1]
		if raw, err := os.ReadFile(path); err == nil {
			fmt.Fprintf(&b, "last %d lines of %s\n%s\n", logFileTailLines, path, lastLines(string(raw), logFileTailLines))
		}
	}
	if strings.TrimSpace(output) != "" {
		b.WriteString(lastLines(output, outputTailLines) + "\n")
	}
	return b.String()
}
