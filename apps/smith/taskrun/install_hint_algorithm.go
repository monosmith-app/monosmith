package taskrun

import "strings"

// installHints maps a program name to where a person gets it.
var installHints = map[string]string{
	"dotnet": "https://dot.net/download", "go": "https://go.dev/dl",
	"node": "https://nodejs.org", "npm": "https://nodejs.org", "npx": "https://nodejs.org",
	"python": "https://python.org/downloads", "python3": "https://python.org/downloads",
	"pip": "https://python.org/downloads", "cargo": "https://rustup.rs", "rustc": "https://rustup.rs",
	"java": "https://adoptium.net", "javac": "https://adoptium.net", "gradle": "https://gradle.org/install",
	"mvn": "https://maven.apache.org/download.cgi", "flutter": "https://docs.flutter.dev/get-started/install",
	"dart": "https://dart.dev/get-dart", "git": "https://git-scm.com/downloads",
	"docker": "https://docs.docker.com/get-docker", "cmake": "https://cmake.org/download",
}

// MissingProgramMessage says a program is not on PATH and how to get it.
//
// Parameters: program is the missing program; installable reports whether
// smith tools install knows a tool by that name (nil means it does not).
// The message names smith tools install only when installable says yes. When
// it says no, the message gives the install page and says smith cannot install
// the program.
func MissingProgramMessage(program string, installable func(name string) bool) string {
	base := strings.TrimSuffix(strings.ToLower(program), ".exe")
	msg := program + " is not installed or not on PATH."
	url, hasPage := installHints[base]
	known := installable != nil && installable(base)
	switch {
	case known && hasPage:
		msg += " Install it from " + url + ", or run smith tools install " + base + " to choose how, then run the command again."
	case known:
		msg += " Run smith tools install " + base + " to choose how, then run the command again."
	case hasPage:
		msg += " smith cannot install it. Install it from " + url + ", then run the command again."
	}
	return msg
}
