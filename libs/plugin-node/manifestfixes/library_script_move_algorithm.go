package manifestfixes

import (
	"regexp"
	"strings"
)

// shellSyntax is what a build script does besides running commands one after
// the other: a pipe, a redirect, a subshell. smith does not parse it.
var shellSyntax = regexp.MustCompile("\\|\\||[|;<>`]|\\$\\(")

// scriptToken is one word of a command, with where it sits in the script.
var scriptToken = regexp.MustCompile(`\S+`)

// libraryTools are the programs of a build script smith knows to write only
// where its flags say: a bundler, the compiler, and smith's own post step.
var libraryTools = map[string]bool{"esbuild": true, "tsc": true, "smith": true}

// moveLibraryScript moves every folder or file a build script writes from the
// old build folder to newDir, in place: the value of --outfile, --outdir,
// --declarationDir and --tsBuildInfoFile, and the folder smith
// declaration-specifiers works in. Every other character stays.
//
// It is conservative. A script it cannot read with certainty is refused, with
// the reason: a command other than esbuild, tsc and smith that names the old
// folder, or shell syntax (a pipe, a redirect, a subshell) in a script that
// names it.
//
// Parameters: script is the text of the script; old the build folder relative
// to the project ("dist"); newDir the folder to write to, relative to the
// project. Returns the new script and the refusal reason ("" unless refused).
func moveLibraryScript(script, old, newDir string) (string, string) {
	if old == "" || old == "." {
		return script, "the build folder is the project folder itself"
	}
	parts := strings.Split(script, " && ")
	if shellSyntax.MatchString(script) && namesFolder(script, old) {
		return script, "the script uses shell syntax (a pipe, a redirect or a subshell) next to " + old + ", which smith does not rewrite"
	}
	for i, part := range parts {
		if !namesFolder(part, old) {
			continue
		}
		program := programOf(part)
		if !libraryTools[program] || (program == "smith" && !strings.Contains(part, "declaration-specifiers")) {
			return script, "the script runs " + program + " on " + old + ", and smith knows only esbuild, tsc and smith declaration-specifiers"
		}
		parts[i] = moveTokens(part, old, newDir)
	}
	return strings.Join(parts, " && "), ""
}

// programOf is the program a command runs, without an npx in front, an
// environment assignment, or a .cmd suffix.
func programOf(command string) string {
	for _, word := range strings.Fields(command) {
		if word == "npx" || strings.Contains(word, "=") {
			continue
		}
		return strings.TrimSuffix(word, ".cmd")
	}
	return ""
}

// namesFolder reports whether any word of the text is the folder or a path
// inside it, alone or as the value of a --flag=value.
func namesFolder(text, old string) bool {
	for _, token := range scriptToken.FindAllString(text, -1) {
		if _, value := splitFlag(token); under(unquote(value), old) {
			return true
		}
	}
	return false
}

// splitFlag splits --flag=value into its flag with the "=" and its value; a
// word that is not a flag with "=" is all value.
func splitFlag(token string) (flag, value string) {
	if strings.HasPrefix(token, "--") {
		if i := strings.Index(token, "="); i > 0 {
			return token[:i+1], token[i+1:]
		}
	}
	return "", token
}

// unquote drops one pair of quotes around a word.
func unquote(s string) string { return strings.Trim(s, `"'`) }

// moveTokens rewrites each word of the command that is, or is inside, the old
// folder.
func moveTokens(command, old, newDir string) string {
	return scriptToken.ReplaceAllStringFunc(command, func(token string) string {
		flag, value := splitFlag(token)
		bare := unquote(value)
		if !under(bare, old) {
			return token
		}
		quote := ""
		if bare != value {
			quote = value[:1]
		}
		return flag + quote + followPath(bare, old, newDir) + quote
	})
}
