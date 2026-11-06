package pipeline

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// probe runs a tool for its version; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Probe(dir, argv) }

var goVersion = regexp.MustCompile(`go(\d+)\.(\d+)`)

// CheckToolVersions compares the installed Go with the version go.work
// requires (older fails: the go command refuses the workspace), and an
// installed golangci-lint with the version CI pins (a difference warns: lint
// results may differ from CI's).
func CheckToolVersions(workspaceRoot, required string, lints bool) []protocol.Finding {
	var out []protocol.Finding
	got, err := probe(workspaceRoot, []string{"go", "env", "GOVERSION"})
	m := goVersion.FindStringSubmatch(got)
	switch {
	case err != nil || m == nil:
		out = append(out, protocol.Finding{Check: "Go is installed", Area: "toolchain", Problem: "go is not on PATH", Detail: "go is not on PATH", Remedy: "install Go " + required + " or newer (https://go.dev/dl/)"})
	default:
		have := m[1] + "." + m[2]
		out = append(out, protocol.Finding{
			Check: "Go is at least " + required + " (go.work)", Area: "toolchain", OK: !older(have, required),
			Problem: olderProblem(have, required, strings.TrimSpace(got)),
			Detail:  "Go " + strings.TrimSpace(got) + " is older than the " + required + " go.work requires", Remedy: "install Go " + required + " or newer",
		})
	}
	if !lints {
		return out
	}
	return append(out, lintFinding(workspaceRoot, strings.TrimPrefix(GolangciLintVersion, "v")))
}

// LintInstallCommand is the command that installs the golangci-lint version CI
// pins: the one command the doctor fix runs, not the whole `smith ci setup`.
func LintInstallCommand() []string {
	return []string{"go", "install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@" + GolangciLintVersion}
}

// older compares major.minor versions.
func older(have, want string) bool {
	h, w := strings.SplitN(have, ".", 3), strings.SplitN(want, ".", 3)
	for i := 0; i < 2; i++ {
		var x, y int
		if i < len(h) {
			x, _ = strconv.Atoi(h[i])
		}
		if i < len(w) {
			y, _ = strconv.Atoi(w[i])
		}
		if x != y {
			return x < y
		}
	}
	return false
}

// olderProblem is the finding's Problem for a Go older than go.work requires;
// a Go that is new enough has none.
func olderProblem(have, required, got string) string {
	if !older(have, required) {
		return ""
	}
	return "Go is " + strings.TrimPrefix(got, "go") + ", go.work requires " + required
}

func firstLine(text, fallback string) string {
	if line, _, _ := strings.Cut(strings.TrimSpace(text), "\n"); line != "" {
		return line
	}
	return fallback
}
