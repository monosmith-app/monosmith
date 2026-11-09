package shellcompletion

import (
	"errors"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"io/fs"
	"strings"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Install writes the files of a plan. A file that is already current stays as
// it is, so a second run changes nothing and adds no second block. A file of
// the person is only appended to, between the markers.
//
// Parameters: env is the machine; plan is what PlanFor returned and the
// person confirmed.
// Returns: what happened to each file. The error says that a file that smith
// did not write is in the way (nothing is written then) or that a write failed.
func Install(env shellenv.Environment, plan Plan) ([]Change, error) {
	if t, blocked := plan.Blocked(); blocked {
		return nil, errors.New(words.Say("completion.error.foreign", "path", t.Path))
	}
	var out []Change
	for _, t := range plan.Targets {
		raw, err := env.ReadFile(t.Path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return out, err
		}
		existing := string(raw)
		next := WithBlock(existing, t.Body)
		switch {
		case err == nil && next == string(raw):
			out = append(out, Change{Path: t.Path, Action: "unchanged"})
			continue
		case err == nil && HasBlock(string(raw)):
			out = append(out, Change{Path: t.Path, Action: "updated"})
		default:
			out = append(out, Change{Path: t.Path, Action: "wrote"})
		}
		if err := env.WriteFile(t.Path, []byte(next)); err != nil {
			return out, err
		}
	}
	return out, nil
}

// Uninstall takes out of the files of a shell what Install put there: the
// block, and with it the whole file when nothing else is left in it. A file
// that holds no block is not touched.
//
// Parameters: env is the machine; shell is the shell.
// Returns: what happened to each file that held a block.
func Uninstall(env shellenv.Environment, shell shellenv.Shell) ([]Change, error) {
	var out []Change
	for _, t := range Targets(env, shell, "") {
		raw, err := env.ReadFile(t.Path)
		if err != nil {
			continue
		}
		rest, found := WithoutBlock(string(raw))
		if !found {
			continue
		}
		if left := strings.TrimSpace(rest); left == "" {
			err = env.Remove(t.Path)
		} else {
			err = env.WriteFile(t.Path, []byte(rest))
		}
		if err != nil {
			return out, err
		}
		out = append(out, Change{Path: t.Path, Action: "removed"})
	}
	return out, nil
}

// Installed lists the shells that hold a block of smith in a target file.
//
// Parameters: env is the machine.
func Installed(env shellenv.Environment) []shellenv.Shell {
	var out []shellenv.Shell
	for _, s := range shellenv.Shells {
		for _, t := range Targets(env, s, "") {
			if StateOf(env, t) == Current {
				out = append(out, s)
				break
			}
		}
	}
	return out
}
