package eslintfix

import (
	"encoding/json"
	"strings"
)

// fileResult is the part of one entry of ESLint's JSON report that is counted.
type fileResult struct {
	ErrorCount   int `json:"errorCount"`
	WarningCount int `json:"warningCount"`
}

// countProblems reads ESLint's JSON report (--format json) and returns the
// errors and warnings it lists. ok is false when the text is not that report.
func countProblems(stdout string) (errors, warnings int, ok bool) {
	var files []fileResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &files); err != nil {
		return 0, 0, false
	}
	for _, f := range files {
		errors += f.ErrorCount
		warnings += f.WarningCount
	}
	return errors, warnings, true
}

// firstErrorLine is the line of a crashed ESLint's output that says what went
// wrong: the first "Error"-like line, past ESLint's "Oops!" banner and its
// version line; else the first line that is not one of those.
func firstErrorLine(output string) string {
	var first string
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "Oops!") || strings.HasPrefix(line, "ESLint:") || strings.HasPrefix(line, "at ") {
			continue
		}
		if strings.Contains(line, "Error") || strings.HasPrefix(line, "Key ") {
			return line
		}
		if first == "" {
			first = line
		}
	}
	return first
}
