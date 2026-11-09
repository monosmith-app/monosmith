package shellcompletion

import (
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"strings"
)

// StateOf says how a target stands on disk.
//
// Parameters: env is the machine; t is the target, with its Body (an empty
// Body makes any block count as Current).
// Returns: Missing when the file or the block is not there, Current or
// Outdated when the block is there, Foreign when a whole file is there that
// holds no block.
func StateOf(env shellenv.Environment, t Target) State {
	raw, err := env.ReadFile(t.Path)
	if err != nil || strings.TrimSpace(string(raw)) == "" {
		return Missing
	}
	body, ok := BlockBody(string(raw))
	switch {
	case !ok && t.Whole:
		return Foreign
	case !ok:
		return Missing
	case t.Body == "" || normalise(body) == normalise(t.Body):
		return Current
	}
	return Outdated
}

// normalise cuts the line ends and the space at the ends, so that a file with
// Windows line ends equals its script.
func normalise(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}

// PlanFor works out what install would do for a shell, and writes nothing.
//
// Parameters: env is the machine; shell is the shell; script is the
// completion script of the shell.
// Returns: the plan. Its Targets are empty when the shell has no known
// location.
func PlanFor(env shellenv.Environment, shell shellenv.Shell, script string) Plan {
	p := Plan{Shell: shell, Targets: Targets(env, shell, script), PackagedAt: PackagedPath(env, shell)}
	for i := range p.Targets {
		p.Targets[i].State = StateOf(env, p.Targets[i])
	}
	return p
}

// Installed is true when every target of the plan holds the script of this
// smith, or the package ships the file.
func (p Plan) Installed() bool {
	if p.PackagedAt != "" {
		return true
	}
	for _, t := range p.Targets {
		if t.State != Current {
			return false
		}
	}
	return len(p.Targets) > 0
}

// Blocked returns the first target that holds a file smith did not write.
func (p Plan) Blocked() (Target, bool) {
	for _, t := range p.Targets {
		if t.State == Foreign {
			return t, true
		}
	}
	return Target{}, false
}
