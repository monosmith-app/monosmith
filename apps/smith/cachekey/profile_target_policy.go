package cachekey

import (
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/shellcompletion"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
)

// The environment variables that the key screens keep for the user.
const (
	// KeyVariable holds the key that signs the remote cache.
	KeyVariable = "SMITH_CACHE_KEY"
	// SASVariable holds the SAS token of an Azure Blob remote cache.
	SASVariable = "AZURE_STORAGE_SAS_TOKEN"
)

// Variables lists every variable that the key screens keep, so that removal
// undoes each of them.
var Variables = []string{KeyVariable, SASVariable}

// MaskedKey stands for the key wherever a line is shown to the person.
const MaskedKey = "••••"

// markersFor are the lines that frame the block that holds a variable in a
// shell profile. The key keeps "smith cache key"; the other variables name
// themselves.
func markersFor(variable string) shellcompletion.Markers {
	name := "smith cache key"
	if variable != KeyVariable {
		name = "smith " + strings.ToLower(variable)
	}
	return shellcompletion.Markers{
		Begin: "# >>> " + name + " >>>",
		End:   "# <<< " + name + " <<<",
		Note:  `# Managed by "smith cache setup". Remove it with "smith cache key --remove".`,
	}
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
// A value with a character that a double quote does not hold safely is put in
// single quotes with the right escapes.
//
// Parameters: variable is the variable; value is the secret, or MaskedKey or a
// placeholder to show the line.
func (t ProfileTarget) Line(variable, value string) string {
	quoted := `"` + value + `"`
	if strings.ContainsAny(value, "\"$`\\!'") {
		if t.Shell == shellenv.Fish {
			quoted = "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(value) + "'"
		} else {
			quoted = "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
		}
	}
	if t.Shell == shellenv.Fish {
		return "set -gx " + variable + " " + quoted
	}
	return "export " + variable + "=" + quoted
}
