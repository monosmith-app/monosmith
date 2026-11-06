package taskoutput

import (
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Spec describes how one build command names the folder it writes.
type Spec struct {
	// Match reports whether the command is the build this spec describes.
	Match func(command []string) bool
	// Flags are the options that name the output ("-o", "--outdir").
	Flags []string
	// Flag is the option the remedy tells the user to pass.
	Flag string
	// Default says where the tool writes when no option is given ("beside
	// the sources").
	Default string
	// Value turns what the user writes after the flag into the remedy's
	// example (the output folder with a file name, or the folder itself).
	Value func(target string) string
}

// Check reads the build task of a project. A command that matches the spec and
// passes no output option, or one whose option names a folder outside the
// workspace's dist/, builds where the tool puts it by default: that is a
// finding, with the manual step, since smith does not rewrite a command line.
// A task smith made always passes the option, so a failing finding means the
// command was written or changed by hand.
//
// The option's value is read only when it is a plain path or starts with
// {workspaceRoot}; a value with another placeholder is left unchecked.
//
// Parameters: p is the project; task the task name; spec the build command.
// Returns the finding, and false when the task is absent, is not that
// command, builds into dist/, or cannot be read.
func Check(p protocol.Project, task string, spec Spec) (protocol.Finding, bool) {
	t, ok := p.Tasks[task]
	if !ok || !spec.Match(t.Command) {
		return protocol.Finding{}, false
	}
	target := protocol.ProjectBuildDir(p.Root)
	check := p.Name + " builds into " + target
	remedy := "pass " + spec.Flag + " " + spec.Value("{workspaceRoot}/"+target) + " in the command of the " + task + " task of [projects." + p.Name + ".tasks." + task + "] in smith.toml"
	for i, arg := range t.Command {
		if !slices.Contains(spec.Flags, arg) && !hasFlagPrefix(arg, spec.Flags) {
			continue
		}
		value := flagValue(t.Command, i, spec.Flags)
		if value == "" || strings.Contains(strings.ReplaceAll(value, "{workspaceRoot}", ""), "{") {
			return protocol.Finding{}, false
		}
		resolved := resolve(p, t, value)
		if protocol.InBuildDir(resolved) {
			return protocol.Finding{}, false
		}
		return protocol.Finding{Check: check, Project: p.Name, Problem: "builds into " + resolved + ", not " + target, Detail: "the " + task + " task writes into " + resolved + ", not into " + target + "/", Remedy: remedy}, true
	}
	return protocol.Finding{Check: check, Project: p.Name, Problem: "builds " + spec.Default + ", not into " + target, Detail: "the " + task + " task (" + strings.Join(t.Command, " ") + ") passes no " + spec.Flag + ", so it writes " + spec.Default + ", not into " + target + "/", Remedy: remedy}, true
}

func hasFlagPrefix(arg string, flags []string) bool {
	for _, f := range flags {
		if strings.HasPrefix(f, "--") && strings.HasPrefix(arg, f+"=") {
			return true
		}
	}
	return false
}

func flagValue(command []string, i int, flags []string) string {
	arg := command[i]
	if slices.Contains(flags, arg) {
		if i+1 < len(command) {
			return command[i+1]
		}
		return ""
	}
	_, v, _ := strings.Cut(arg, "=")
	return v
}

// resolve is the folder an option names, from the workspace root.
func resolve(p protocol.Project, t protocol.Task, value string) string {
	if rest, ok := strings.CutPrefix(value, "{workspaceRoot}"); ok {
		return path.Clean(strings.TrimPrefix(rest, "/"))
	}
	if filepath.IsAbs(value) {
		return value
	}
	base := t.Cwd
	if base == "" {
		base = p.Root
	}
	return path.Clean(path.Join(base, value))
}
