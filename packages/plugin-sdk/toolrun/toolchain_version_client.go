package toolrun

import (
	"regexp"
	"strings"
	"sync"
)

var versionNumber = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?(?:[-+.][0-9A-Za-z.]+)*`)

// toolchains remembers what each probe answered, so a plugin that describes
// many projects runs a tool's version once per process.
var toolchains sync.Map

// Toolchain names the version of the tool a project's tasks run with, as
// "node 20.11.0": label, then the first version number the command prints
// (stdout and stderr both: java and older dart print it to stderr). It is ""
// when the tool is missing or prints no version, and a plugin then leaves the
// toolchain out of its project. The answer is probed once per process.
func Toolchain(label string, argv ...string) string {
	key := label + "\x00" + strings.Join(argv, "\x00")
	if v, ok := toolchains.Load(key); ok {
		return v.(string)
	}
	answer := ""
	if out, errOut, err := Capture("", argv); err == nil {
		if v := versionNumber.FindString(firstLine(out + "\n" + errOut)); v != "" {
			answer = label + " " + v
		}
	}
	toolchains.Store(key, answer)
	return answer
}

// firstLine is the first non-blank line of text.
func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			return line
		}
	}
	return ""
}
