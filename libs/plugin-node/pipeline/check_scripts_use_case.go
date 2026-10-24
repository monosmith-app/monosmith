package pipeline

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// scriptsThatMatter are the scripts whose failure is a broken workspace, not
// a leftover: a problem in one makes the finding a failure, else a warning.
var scriptsThatMatter = map[string]bool{
	"build": true, "test": true, "lint": true, "typecheck": true,
	"prebuild": true, "postbuild": true, "pretest": true, "start": true, "dev": true,
}

// knownTools run without being installed in node_modules: they come with Node
// or npm, or are smith itself.
var knownTools = map[string]bool{"node": true, "npm": true, "npx": true, "smith": true, "corepack": true}

// shellCommands are commands the shell a script runs in provides.
var shellCommands = map[string]bool{
	"cd": true, "echo": true, "exit": true, "export": true, "set": true, "test": true, "true": true, "false": true,
	"rm": true, "cp": true, "mv": true, "mkdir": true, "cat": true, "ls": true, "touch": true, "sleep": true,
	"printf": true, "pwd": true, "if": true, "for": true, "while": true, "source": true, ".": true, ":": true,
	"del": true, "copy": true, "move": true, "rmdir": true, "type": true, "call": true, "start": true, "git": true,
}

// CheckScripts checks that the package.json in dir (a folder relative to the
// workspace root; empty for the root) has scripts that can run: each file a
// script hands to node, or names after --config or -c, exists, and each tool
// a script starts is installed (in a node_modules/.bin, or on the PATH). The
// tools are judged only once dependencies are installed, since before that
// none are. label names the package in the finding.
func CheckScripts(workspaceRoot, dir, label string) (protocol.Finding, bool) {
	pkgDir := filepath.Join(workspaceRoot, filepath.FromSlash(dir))
	raw, err := os.ReadFile(filepath.Join(pkgDir, "package.json"))
	if err != nil {
		return protocol.Finding{}, false
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(raw, &pkg) != nil || len(pkg.Scripts) == 0 {
		return protocol.Finding{}, false
	}
	_, installedErr := os.Stat(filepath.Join(workspaceRoot, "node_modules"))
	installed := installedErr == nil

	names := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	var problems, tools []string
	failing, anyFile := false, false
	for _, name := range names {
		files, missingTools := scriptProblems(workspaceRoot, pkgDir, pkg.Scripts[name], installed)
		for _, f := range files {
			problems = append(problems, name+" runs "+f+", which does not exist")
			anyFile = true
		}
		for _, tool := range missingTools {
			problems = append(problems, name+" starts "+tool+", which is not installed")
			if !contains(tools, tool) {
				tools = append(tools, tool)
			}
		}
		if (len(files) > 0 || len(missingTools) > 0) && scriptsThatMatter[name] {
			failing = true
		}
	}
	finding := protocol.Finding{Check: "the scripts of " + label + " call files and tools that exist", OK: len(problems) == 0}
	if len(problems) == 0 {
		return finding, true
	}
	finding.Warning = !failing
	finding.Detail = strings.Join(problems, "; ")
	var remedy []string
	if anyFile {
		remedy = append(remedy, "restore the file the script runs, or change or remove that script")
	}
	if len(tools) > 0 {
		remedy = append(remedy, "install the missing tool (npm i -D "+strings.Join(tools, " ")+" at the workspace root), or remove the script that starts it")
	}
	finding.Remedy = strings.Join(remedy, "\n")
	return finding, true
}

// scriptProblems returns the files a script command runs that are missing and
// the tools it starts that are not installed.
func scriptProblems(workspaceRoot, pkgDir, command string, installed bool) (files, tools []string) {
	for _, segment := range splitCommands(command) {
		words := strings.Fields(segment)
		for len(words) > 0 && strings.Contains(words[0], "=") && !strings.HasPrefix(words[0], "-") {
			words = words[1:]
		}
		if len(words) == 0 {
			continue
		}
		tool := words[0]
		switch {
		case !plainWord(tool) || knownTools[tool] || shellCommands[tool]:
		case strings.Contains(tool, "/"):
			if !exists(pkgDir, tool) {
				files = append(files, tool)
			}
		case installed && !toolInstalled(workspaceRoot, pkgDir, tool):
			tools = append(tools, tool)
		}
		if tool == "node" {
			if script := nodeScript(words[1:]); script != "" && !exists(pkgDir, script) {
				files = append(files, "node "+script)
			}
		}
		for i := 1; i+1 < len(words); i++ {
			if (words[i] == "-c" || words[i] == "--config") && plainWord(words[i+1]) && looksLikeFile(words[i+1]) && !exists(pkgDir, words[i+1]) {
				files = append(files, tool+" "+words[i]+" "+words[i+1])
			}
		}
	}
	return files, tools
}

// splitCommands splits a script at the shell operators that start another
// command.
func splitCommands(script string) []string {
	r := strings.NewReplacer("&&", "\n", "||", "\n", ";", "\n", "|", "\n")
	return strings.Split(r.Replace(script), "\n")
}

// plainWord is a word with nothing the shell would expand or a program would
// compute: a literal command or path.
func plainWord(w string) bool {
	return w != "" && !strings.ContainsAny(w, "$*?{}()<>`\"'\\!~") && !strings.HasPrefix(w, "-")
}

func looksLikeFile(w string) bool {
	return strings.ContainsAny(w, "/.") && !strings.HasPrefix(w, "http")
}

// nodeScript is the file node is asked to run: the first argument that is not
// an option. An inline program (-e, -p) has none.
func nodeScript(args []string) string {
	withValue := map[string]bool{"-r": true, "--require": true, "--import": true, "--loader": true, "--experimental-loader": true, "--env-file": true}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-e" || a == "-p" || a == "--eval" || a == "--print" || a == "--test":
			return ""
		case withValue[a]:
			i++
		case strings.HasPrefix(a, "-"):
		case plainWord(a):
			return a
		default:
			return ""
		}
	}
	return ""
}

func exists(dir, rel string) bool {
	_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel)))
	return err == nil
}

// toolInstalled reports whether a program a script starts can be found: in a
// node_modules/.bin from the package up to the workspace root, or on the PATH.
func toolInstalled(workspaceRoot, pkgDir, tool string) bool {
	for dir := pkgDir; ; dir = filepath.Dir(dir) {
		for _, name := range []string{tool, tool + ".cmd", tool + ".ps1"} {
			if _, err := os.Stat(filepath.Join(dir, "node_modules", ".bin", name)); err == nil {
				return true
			}
		}
		if dir == workspaceRoot || filepath.Dir(dir) == dir {
			break
		}
	}
	_, err := exec.LookPath(tool)
	return err == nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
