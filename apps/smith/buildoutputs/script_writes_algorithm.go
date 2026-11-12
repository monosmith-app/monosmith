package buildoutputs

import (
	"path"
	"slices"
	"strings"
)

// configTools take the folder they write from a config file when their
// command line names none, so the command line alone cannot tell it.
var configTools = map[string]bool{"tsc": true, "rollup": true, "webpack": true, "tsup": true, "ng": true, "next": true}

// RunScript reads a command that runs a package.json script with npm: `npm
// run <script>`, with an optional `-w <package>` (or --workspace) and
// --silent or --if-present.
//
// Parameters: command is the task's argv.
// Returns: the script's name, the package -w names ("" when none), and false
// when the command is anything else (extra arguments included, since they
// change what the script runs).
func RunScript(command []string) (script, pkg string, ok bool) {
	if len(command) < 3 || command[0] != "npm" || (command[1] != "run" && command[1] != "run-script") {
		return "", "", false
	}
	rest := command[3:]
	for i := 0; i < len(rest); i++ {
		switch a := rest[i]; {
		case a == "-w" || a == "--workspace":
			if i+1 >= len(rest) {
				return "", "", false
			}
			pkg = rest[i+1]
			i++
		case strings.HasPrefix(a, "--workspace="):
			pkg = strings.TrimPrefix(a, "--workspace=")
		case a == "--silent" || a == "-s" || a == "--if-present":
		default:
			return "", "", false
		}
	}
	return command[2], pkg, true
}

// ExpandScript replaces each `npm run <name>` step of a script with the
// script of that name, so what the steps write can be read. A name the
// scripts do not hold, a step with extra arguments and a script that runs
// itself again are left as they are.
//
// Parameters: script is the script text; scripts are the package.json
// scripts of the same package.
// Returns: the script with its npm run steps expanded, one step per line.
func ExpandScript(script string, scripts map[string]string) string {
	return expand(script, scripts, map[string]bool{})
}

func expand(script string, scripts map[string]string, seen map[string]bool) string {
	var steps []string
	for _, segment := range strings.Split(shellSplit.Replace(script), "\n") {
		words := strings.Fields(segment)
		if name, pkg, ok := RunScript(words); ok && pkg == "" && scripts[name] != "" && !seen[name] {
			seen[name] = true
			steps = append(steps, expand(scripts[name], scripts, seen))
			delete(seen, name)
			continue
		}
		if len(words) > 0 {
			steps = append(steps, strings.Join(words, " "))
		}
	}
	return strings.Join(steps, "\n")
}

// ScriptWrites returns the directories, relative to the project, a build
// script writes, as far as its command line tells: what WrittenDirs reads,
// plus dist for a `vite build` that passes no output folder (Vite's default),
// unless the project's Vite config names one.
//
// Parameters: script is the build script, with its npm run steps expanded;
// viteNamesOutDir tells whether the project's vite.config names an outDir.
// Returns: the directories, and complete = false when a step runs a tool that
// reads its output folder from a config file (tsc without --outDir, rollup -c
// without -o, a Vite config with an outDir): what it writes is then unknown.
func ScriptWrites(script string, viteNamesOutDir bool) (dirs []string, complete bool) {
	dirs = WrittenDirs(script)
	complete = true
	for _, segment := range strings.Split(shellSplit.Replace(script), "\n") {
		words := strings.Fields(segment)
		if len(words) > 0 && words[0] == "npx" {
			words = words[1:]
		}
		if len(words) == 0 || len(WrittenDirs(segment)) > 0 || namesOutput(words) {
			continue
		}
		switch {
		case words[0] == "vite" && len(words) > 1 && words[1] == "build":
			if viteNamesOutDir {
				complete = false
			} else if !slices.Contains(dirs, "dist") {
				dirs = append(dirs, "dist")
			}
		case configTools[words[0]] && !slices.Contains(words, "--noEmit"):
			complete = false
		}
	}
	return dirs, complete
}

// namesOutput reports whether a step passes an output option, even one whose
// value WrittenDirs leaves out (a shell variable).
func namesOutput(words []string) bool {
	for _, w := range words {
		flag, _, _ := strings.Cut(w, "=")
		if dirFlags[flag] || fileFlags[flag] || eitherFlags[flag] {
			return true
		}
	}
	return false
}

// Suggested returns the outputs a build task should declare: each directory
// the build writes, or the declared output that holds it, so a parent folder
// the task already declares is kept. A declared output nothing is written to
// is left out.
//
// Parameters: written are the directories the build writes; declared the
// task's outputs, both relative to the project.
// Returns: the outputs, outermost only, in the order written.
func Suggested(written, declared []string) []string {
	var out []string
	for _, w := range Collapse(written) {
		pick := path.Clean(w)
		for _, d := range declared {
			if Covered(w, []string{d}) {
				pick = path.Clean(d)
				break
			}
		}
		if !slices.Contains(out, pick) {
			out = append(out, pick)
		}
	}
	return Collapse(out)
}
