package pipeline

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// probe runs a tool for its version; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Probe(dir, argv) }

// CheckNodeVersion compares the installed Node.js with the major CI
// installs. Another major, newer or older, is a warning: a team may run it on
// purpose, and the finding says what differs and why it can matter. The fix
// is a question with the choices: keep it (nothing changes, and the first
// answer), or get the pinned Node.js with smith tools install, which asks how
// and runs nothing before the person picks. Node.js missing from the PATH is a
// failure.
func CheckNodeVersion(workspaceRoot, pin string) protocol.Finding {
	f := protocol.Finding{Check: "Node.js is " + pin + ", as in CI", Area: "toolchain", Warning: true, Remedy: "install Node.js " + pin + " (https://nodejs.org)"}
	got, err := probe(workspaceRoot, []string{"node", "--version"})
	if err != nil {
		f.Warning, f.Problem, f.Detail = false, "Node.js is not on PATH, CI installs "+pin, "node is not on PATH"
		return f
	}
	have := strings.TrimPrefix(strings.TrimSpace(got), "v")
	if found, ok := toolrun.CIPinFinding("Node.js", have, pin); ok {
		found.Remedy = f.Remedy
		if !found.OK {
			found.Detail += ". Nothing is broken by this: a difference is a warning, and the choices are below"
			found.Fix = pinFix(have, pin)
		}
		return found
	}
	f.Detail = "found Node.js " + have
	return f
}

// pinFix asks what to do about a Node.js that is not the major CI installs.
// Keeping it is the first answer and changes nothing. The other asks smith
// tools install, which shows the install commands and runs only the one the
// person picks. No setting moves the pin CI installs, so there is no third
// answer that edits one.
func pinFix(have, pin string) *protocol.Fix {
	keep := "Keep Node.js " + have + " as it is: nothing changes, and the difference stays a warning"
	install := "Get Node.js " + pin + " with smith tools install (smith asks how, and nothing runs before you choose)"

	return &protocol.Fix{
		Description: "Decide what to do about Node.js " + have + " (CI installs " + pin + ")",
		Choices: []protocol.FixChoice{{
			Question: "Node.js " + have + " runs here and CI installs " + pin + ". Builds can differ between this machine and CI. What do you want to do?",
			Options: []protocol.FixChoiceOption{
				{Label: keep, Fix: protocol.Fix{Description: keep}},
				{Label: install, Fix: protocol.Fix{Description: install, Steps: []protocol.FixStep{{Description: "smith tools install node: it shows the commands and asks before it runs one", Argv: []string{"smith", "tools", "install", "node"}}}}},
			},
		}},
	}
}
