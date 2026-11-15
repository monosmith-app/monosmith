package lintfix

import (
	"encoding/json"
	"fmt"
	"strings"
)

// fileResult is the part of one entry of ESLint's JSON report that is counted.
type fileResult struct {
	ErrorCount   int `json:"errorCount"`
	WarningCount int `json:"warningCount"`
}

// ReadESLint reads what an ESLint run printed (--format json) and returns the
// errors and warnings its report lists. A run that exits above 1, or prints
// no report, crashed: crash is then the line that says why. Exit 1 means the
// report lists problems.
//
// Parameters: stdout and stderr are the run's two outputs, code its exit code.
func ReadESLint(stdout, stderr string, code int) (errors, warnings int, crash string) {
	var files []fileResult
	parsed := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &files) == nil
	if code > 1 || !parsed {
		if crash = firstErrorLine(stderr + "\n" + stdout); crash == "" {
			crash = fmt.Sprintf("ESLint exited with code %d", code)
		}
		return 0, 0, crash
	}
	for _, f := range files {
		errors += f.ErrorCount
		warnings += f.WarningCount
	}
	return errors, warnings, ""
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
