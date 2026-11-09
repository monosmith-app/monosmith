package shellcompletion

import (
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"path/filepath"
)

// PackagedPath finds the completion file that a package of smith (Homebrew, a
// Linux package) already ships for a shell. A file that holds the marker lines
// is smith's own install and does not count.
//
// Parameters: env is the machine; shell is the shell.
// Returns: the path of the file, and "" when the system has none.
func PackagedPath(env shellenv.Environment, shell shellenv.Shell) string {
	for _, p := range packagedCandidates(env, shell) {
		if !env.Exists(p) {
			continue
		}
		if raw, err := env.ReadFile(p); err == nil && HasBlock(string(raw)) {
			continue
		}
		return p
	}
	return ""
}

// packagedCandidates are the places a package puts the completion file of
// smith. They are the system folders that bash, zsh and fish read.
func packagedCandidates(env shellenv.Environment, shell shellenv.Shell) []string {
	if env.OS() == "windows" {
		return nil
	}
	prefixes := []string{"/usr", "/usr/local", "/opt/homebrew", "/home/linuxbrew/.linuxbrew"}
	if p := env.Getenv("HOMEBREW_PREFIX"); p != "" {
		prefixes = append([]string{p}, prefixes...)
	}
	var out []string
	for _, p := range prefixes {
		switch shell {
		case shellenv.Bash:
			out = append(out, filepath.Join(p, "share", "bash-completion", "completions", "smith"), filepath.Join(p, "etc", "bash_completion.d", "smith"))
		case shellenv.Zsh:
			out = append(out, filepath.Join(p, "share", "zsh", "site-functions", "_smith"))
		case shellenv.Fish:
			out = append(out, filepath.Join(p, "share", "fish", "vendor_completions.d", "smith.fish"))
		}
	}
	if shell == shellenv.Bash {
		out = append(out, "/etc/bash_completion.d/smith")
	}
	return out
}
