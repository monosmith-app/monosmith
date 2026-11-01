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

// MissingProgramMessage says a program is not on PATH and, when smith knows
// where to get it, adds the address to install it from.
func MissingProgramMessage(program string) string {
	base := strings.TrimSuffix(strings.ToLower(program), ".exe")
	msg := program + " is not installed or not on PATH."
	if url, ok := installHints[base]; ok {
		msg += " Install it from " + url + ", then run the command again."
	}
	return msg
}
