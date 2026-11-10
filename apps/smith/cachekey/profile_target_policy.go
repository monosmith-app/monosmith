package cachekey

import (
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/shellcompletion"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
)

// KeyVariable is the environment variable that holds the key.
const KeyVariable = "SMITH_CACHE_KEY"

// MaskedKey stands for the key wherever a line is shown to the person.
const MaskedKey = "••••"

// profileMarkers frame the block that holds the key in a shell profile.
var profileMarkers = shellcompletion.Markers{
	Begin: "# >>> smith cache key >>>",
	End:   "# <<< smith cache key <<<",
	Note:  `# Managed by "smith cache key". Remove it with "smith cache key --remove".`,
}

// ProfileTarget is the file of a shell where the key can go.
type ProfileTarget struct {
	// Shell is the shell that reads the file.
	Shell shellenv.Shell
	// Path is the file.
	Path string
}

// ProfileTargets lists the file of every shell that can hold the key on this
// machine, so that removal finds the block whatever shell wrote it.
//
// Parameters: env is the machine.
// Returns: the targets; empty when no home folder is known.
func ProfileTargets(env shellenv.Environment) []ProfileTarget {
	home := env.HomeDir()
	if home == "" {
		return nil
	}
	rc := env.Getenv("ZDOTDIR")
	if rc == "" {
		rc = home
	}
	conf := env.Getenv("XDG_CONFIG_HOME")
	if conf == "" {
		conf = filepath.Join(home, ".config")
	}
	bash := ".bashrc"
	if env.OS() == "darwin" {
		bash = ".bash_profile"
	}
	return []ProfileTarget{
		{Shell: shellenv.Zsh, Path: filepath.Join(rc, ".zshrc")},
		{Shell: shellenv.Bash, Path: filepath.Join(home, bash)},
		{Shell: shellenv.Fish, Path: filepath.Join(conf, "fish", "conf.d", "smith-cache.fish")},
	}
}

// TargetFor picks the file of the shell the person works in.
//
// Parameters: env is the machine.
// Returns: the target, and false when the shell is not zsh, bash or fish, or
// when no home folder is known. The caller then shows the line to add by hand.
func TargetFor(env shellenv.Environment) (ProfileTarget, bool) {
	shell, ok := shellenv.DetectShell(env)
	if !ok {
		return ProfileTarget{}, false
	}
	for _, t := range ProfileTargets(env) {
		if t.Shell == shell {
			return t, true
		}
	}
	return ProfileTarget{}, false
}

// Line is the line that sets the variable, in the syntax of the shell.
//
// Parameters: value is the key, or MaskedKey or a placeholder to show the line.
func (t ProfileTarget) Line(value string) string {
	if t.Shell == shellenv.Fish {
		return "set -gx " + KeyVariable + ` "` + value + `"`
	}
	return "export " + KeyVariable + `="` + value + `"`
}
