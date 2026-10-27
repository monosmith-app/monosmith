package cipipeline

import (
	"regexp"
	"strings"
)

const (
	// tailKeep bounds what a step's captured output holds in memory.
	tailKeep        = 64 * 1024
	outputTailLines = 30
	logFileLines    = 40
)

// tailWriter keeps the end of what is written to it, so a failed step's last
// lines can be shown again while the step's output still streams as it comes.
type tailWriter struct{ buf []byte }

func (t *tailWriter) Write(p []byte) (int, error) {
	t.buf = append(t.buf, p...)
	if len(t.buf) > 2*tailKeep {
		t.buf = append([]byte(nil), t.buf[len(t.buf)-tailKeep:]...)
	}
	return len(p), nil
}

func (t *tailWriter) String() string { return string(t.buf) }

// lastLines is the last n lines of text.
func lastLines(text string, n int) string {
	lines := strings.Split(strings.TrimRight(strings.ReplaceAll(text, "\r\n", "\n"), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

var namedLog = regexp.MustCompile(`(?m)complete log of this run can be found in:\s*(\S.*?)\s*$`)

// logFileNamed is the last log file path the output names (npm prints
// "A complete log of this run can be found in: <path>"), or "".
func logFileNamed(output string) string {
	all := namedLog.FindAllStringSubmatch(output, -1)
	if len(all) == 0 {
		return ""
	}
	return all[len(all)-1][1]
}
