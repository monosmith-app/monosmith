package pipeline

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// setupFix is the fix of a missing venv: `smith ci setup`, which the host runs
// with the smith that is running, after the person agreed.
//
// Parameters: missing is true when the venv is missing.
// Returns: the fix, or nil when the venv exists. It is never Safe: it creates
// the venv and installs the requirements from the network.
func setupFix(missing bool) *protocol.Fix {
	if !missing {
		return nil
	}
	return &protocol.Fix{
		Description: "Create the Python venv and install the requirements",
		Steps: []protocol.FixStep{{
			Description: "Create .venv at the workspace root and install requirements-dev.txt into it (uses the network)",
			Argv:        []string{"smith", "ci", "setup"},
		}},
	}
}
