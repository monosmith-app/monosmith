package pipeline

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// checkIgnore runs git check-ignore for the venv and returns its exit code
// (0 ignored, 1 not ignored, anything else: not a repository or no git);
// tests replace it.
var checkIgnore = func(workspaceRoot string) int {
	_, _, err := toolrun.Capture(workspaceRoot, []string{"git", "check-ignore", "-q", ".venv"})
	var exit *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &exit):
		return exit.ExitCode()
	default:
		return -1
	}
}

// CheckVenvIgnored reports a venv that git would pick up: an un-ignored
// .venv is one `git add .` away from a commit of thousands of files. It
// returns false when there is nothing to say (no venv, or not a git
// repository).
func CheckVenvIgnored(workspaceRoot string) (protocol.Finding, bool) {
	if _, err := os.Stat(filepath.Join(workspaceRoot, ".venv")); err != nil {
		return protocol.Finding{}, false
	}
	code := checkIgnore(workspaceRoot)
	if code != 0 && code != 1 {
		return protocol.Finding{}, false
	}
	f := protocol.Finding{Check: ".venv is ignored by git", OK: code == 0}
	if f.OK {
		return f, true
	}
	f.Detail = ".venv is not in .gitignore, so git add . would commit the whole environment"
	f.Remedy = "add `" + venv.IgnoreEntry + "` to .gitignore (smith doctor --fix does)"
	current, _ := os.ReadFile(filepath.Join(workspaceRoot, ".gitignore"))
	if out, changed := venv.GitignoreWith(string(current)); changed {
		f.Fix = &protocol.Fix{
			Description: "Ignore .venv in .gitignore",
			Safe:        true,
			Files:       []protocol.FileEdit{{Path: ".gitignore", Content: out}},
		}
	}
	return f, true
}
