package commands

import "strings"

// RunRequest is what a `smith run` style command line asks for.
type RunRequest struct {
	Tasks    []string
	Projects []string
	Affected bool
	Base     string
}

var taskShortcuts = map[string]bool{"build": true, "test": true, "lint": true, "typecheck": true}

// valueFlags are the run flags that take a value.
var valueFlags = map[string]bool{"-p": true, "--projects": true, "--base": true, "--parallel": true}

// ParseRun reads a command line (without the program name) of smith run, or
// of build, test, lint or typecheck, so the shell can tell which tasks it will
// start before they do. ok is false for any other command.
func ParseRun(argv []string) (req RunRequest, ok bool) {
	if len(argv) == 0 {
		return req, false
	}
	verb := argv[0]
	if verb != "run" && !taskShortcuts[verb] {
		return req, false
	}
	var positional []string
	for i := 1; i < len(argv); i++ {
		a := argv[i]
		switch {
		case a == "--affected":
			req.Affected = true
		case a == "-p" || a == "--projects":
			if i+1 < len(argv) {
				i++
				req.Projects = append(req.Projects, strings.Split(argv[i], ",")...)
			}
		case strings.HasPrefix(a, "--projects="):
			req.Projects = append(req.Projects, strings.Split(strings.TrimPrefix(a, "--projects="), ",")...)
		case a == "--base":
			if i+1 < len(argv) {
				i++
				req.Base = argv[i]
			}
		case strings.HasPrefix(a, "--base="):
			req.Base = strings.TrimPrefix(a, "--base=")
		case valueFlags[a]:
			i++
		case strings.HasPrefix(a, "-"):
		default:
			positional = append(positional, a)
		}
	}
	if verb == "run" {
		req.Tasks = positional
	} else {
		req.Tasks = []string{verb}
		req.Projects = append(req.Projects, positional...)
	}
	return req, len(req.Tasks) > 0
}

// RunArgv builds the command line for a task on a project (or all projects
// when project is empty), the way the buttons and shortcuts start one.
func RunArgv(task, project string) []string {
	argv := []string{"run", task}
	if project != "" {
		argv = append(argv, "-p", project)
	}
	return argv
}
