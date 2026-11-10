package shellcompletion

import (
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
)

// Targets lists the files install writes for a shell.
//
// The script never goes into the profile of the person. Smith owns one file
// per shell in its own folder, <smith home>/completions (SMITH_HOME, or
// ~/.smith, like the rest of the per-user state of smith): smith.bash,
// smith.zsh and smith.ps1. The profile (.bashrc, .zshrc, the PowerShell
// profiles) gets one marked line that sources that file when it exists, so a
// deleted file never breaks the start of the shell. Fish loads the script from
// its own completions folder, so it needs no line.
//
// Parameters: env is the machine; shell is the shell; script is the
// completion script (empty when only the paths matter, as in uninstall).
// Returns: the targets in the order they are written: the script file first,
// then the profile lines. It is empty when no home folder is known.
func Targets(env shellenv.Environment, shell shellenv.Shell, script string) []Target {
	home := env.HomeDir()
	if home == "" {
		return nil
	}
	smithHome := orDefault(env.Getenv("SMITH_HOME"), filepath.Join(home, ".smith"))
	dir := filepath.Join(smithHome, "completions")
	switch shell {
	case shellenv.Fish:
		conf := orDefault(env.Getenv("XDG_CONFIG_HOME"), filepath.Join(home, ".config"))
		return []Target{{Path: filepath.Join(conf, "fish", "completions", "smith.fish"), Whole: true, Body: script}}
	case shellenv.Bash:
		file := filepath.Join(dir, "smith.bash")
		return []Target{
			{Path: file, Whole: true, Body: script},
			{Path: filepath.Join(home, ".bashrc"), Body: bashLine(file)},
		}
	case shellenv.Zsh:
		file := filepath.Join(dir, "smith.zsh")
		rc := orDefault(env.Getenv("ZDOTDIR"), home)
		return []Target{
			{Path: file, Whole: true, Body: script},
			{Path: filepath.Join(rc, ".zshrc"), Body: zshLine(file)},
		}
	case shellenv.PowerShell:
		file := filepath.Join(dir, "smith.ps1")
		out := []Target{{Path: file, Whole: true, Body: script}}
		for _, p := range powerShellProfiles(env, home) {
			out = append(out, Target{Path: p, Body: powerShellLine(file)})
		}
		return out
	}
	return nil
}

// orDefault is value, or def when value is empty.
func orDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

// bashLine is the line of .bashrc: it sources the script when the file is there.
func bashLine(file string) string {
	f := `"` + filepath.ToSlash(file) + `"`
	return "[ -f " + f + " ] && source " + f
}

// zshLine is the line of .zshrc. The script needs compdef, so the line loads
// compinit first when the rest of .zshrc did not.
func zshLine(file string) string {
	f := `"` + filepath.ToSlash(file) + `"`
	return "[ -f " + f + " ] && { (( $+functions[compdef] )) || { autoload -Uz compinit && compinit; }; source " + f + "; }"
}

// powerShellLine is the line of a profile: it runs the script when the file is there.
func powerShellLine(file string) string {
	f := "'" + strings.ReplaceAll(file, "'", "''") + "'"
	return "if (Test-Path " + f + ") { . " + f + " }"
}

// powerShellProfiles are the profile files of the current user and current
// host. On Windows Windows PowerShell 5.1 and PowerShell 7 each have one. When
// the parent program says which of them runs, only that one is listed; when it
// does not, both are, and the plan shows them so the person can decline.
func powerShellProfiles(env shellenv.Environment, home string) []string {
	const name = "Microsoft.PowerShell_profile.ps1"
	if env.OS() != "windows" {
		return []string{filepath.Join(orDefault(env.Getenv("XDG_CONFIG_HOME"), filepath.Join(home, ".config")), "powershell", name)}
	}
	docs := filepath.Join(home, "Documents")
	if redirected := filepath.Join(home, "OneDrive", "Documents"); !env.Exists(docs) && env.Exists(redirected) {
		docs = redirected
	}
	v5 := filepath.Join(docs, "WindowsPowerShell", name)
	v7 := filepath.Join(docs, "PowerShell", name)
	switch parent := strings.TrimSuffix(strings.ToLower(env.ParentProgram()), ".exe"); parent {
	case "pwsh":
		return []string{v7}
	case "powershell":
		return []string{v5}
	}
	return []string{v5, v7}
}
