package turboplan

import (
	"strconv"
	"strings"
)

// turboSubcommands are the commands of the turbo CLI; any other first word is
// a task (`turbo build` is `turbo run build`).
var turboSubcommands = map[string]bool{
	"run": true, "watch": true, "prune": true, "gen": true, "generate": true, "login": true,
	"logout": true, "link": true, "unlink": true, "daemon": true, "info": true, "bin": true,
	"telemetry": true, "ls": true, "query": true, "boundaries": true, "scan": true, "devtools": true,
}

// turboValueFlags take a value, in `--flag value` or `--flag=value` form.
var turboValueFlags = map[string]bool{
	"filter": true, "F": true, "concurrency": true, "output-logs": true, "log-order": true,
	"log-prefix": true, "ui": true, "env-mode": true, "cache-dir": true, "token": true, "team": true,
	"api": true, "login": true, "cache": true, "cwd": true, "root-turbo-json": true, "profile": true,
	"anon-profile": true, "since": true, "scope": true, "cache-workers": true, "verbosity": true,
}

// turboQuiet are flags that only tune turbo's own output, daemon or update
// check; smith has nothing to map them to, and nothing to lose.
var turboQuiet = map[string]bool{
	"output-logs": true, "log-order": true, "log-prefix": true, "ui": true, "env-mode": true,
	"summarize": true, "no-daemon": true, "daemon": true, "no-update-notifier": true,
	"framework-inference": true, "color": true, "no-color": true, "verbosity": true, "v": true,
	"verbose": true, "cache-workers": true, "heap": true, "dangerously-disable-package-manager-check": true,
}

// turboNoEquivalent are flags that change what runs or where it is cached,
// which smith has no counterpart for; the reason goes to the report.
var turboNoEquivalent = map[string]string{
	"parallel":        "--parallel ignores the order of dependencies, which smith run always keeps",
	"only":            "--only runs a task without its dependencies, which smith run cannot",
	"dry":             "--dry prints the plan; use smith explain <project>:<task>",
	"graph":           "--graph draws the task graph; use smith graph",
	"cache-dir":       "--cache-dir sets turbo's cache directory; smith caches in .smith/cache",
	"cache":           "--cache sets turbo's cache read and write modes; smith has --no-cache only",
	"token":           "--token is turbo's remote cache login; configure smith's remote cache instead",
	"team":            "--team is turbo's remote cache team; configure smith's remote cache instead",
	"api":             "--api is turbo's remote cache URL; configure smith's remote cache instead",
	"login":           "--login is turbo's remote cache URL; configure smith's remote cache instead",
	"cwd":             "--cwd changes the directory turbo runs in",
	"root-turbo-json": "--root-turbo-json points at another turbo.json",
	"since":           "--since selects packages changed since a commit; use --affected --base",
	"scope":           "--scope is turbo 1's --filter",
	"profile":         "--profile writes a trace",
	"anon-profile":    "--anon-profile writes a trace",
}

// rewriteTurbo turns the arguments of one turbo command into a `smith run`.
func rewriteTurbo(args []string, names map[string]string) (string, []string, string) {
	if len(args) == 0 {
		return "", nil, "turbo with no task"
	}
	if first := args[0]; turboSubcommands[first] && !strings.HasPrefix(first, "-") {
		switch first {
		case "run":
			args = args[1:]
		case "watch":
			return "", nil, "turbo watch reruns tasks when files change, which smith run does not; smith dev starts servers"
		case "prune":
			return "", nil, "turbo prune writes a pruned monorepo for a Dockerfile; for a Node app, smith run prune -p <app> writes its deploy output (production package.json and lockfile)"
		default:
			return "", nil, "turbo " + first + " has no smith equivalent"
		}
	}
	var tasks, filters, flags, dropped []string
	bail := true
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			return "", nil, "it passes arguments to the task (" + strings.Join(args[i:], " ") + "), and smith run does not pass arguments on; put them in the package.json script"
		}
		if !strings.HasPrefix(a, "-") {
			tasks = append(tasks, a)
			continue
		}
		name, value, hasValue := strings.Cut(strings.TrimLeft(a, "-"), "=")
		if !hasValue && turboValueFlags[name] && i+1 < len(args) {
			i++
			value, hasValue = args[i], true
		}
		switch {
		case name == "filter" || name == "F":
			v, problem := filterOf(value, names)
			if problem != "" {
				return "", nil, problem
			}
			filters = append(filters, v)
		case name == "affected":
			flags = append(flags, "--affected")
		case name == "force" || name == "no-cache":
			flags = append(flags, "--no-cache")
		case name == "concurrency":
			if _, err := strconv.Atoi(value); err != nil {
				return "", nil, "--concurrency=" + value + " is not a number of tasks (smith run --parallel takes one)"
			}
			flags = append(flags, "--parallel "+value)
		case name == "continue":
			if hasValue && value != "always" && value != "true" {
				return "", nil, "--continue=" + value + " has no smith equivalent (smith run keeps going unless --bail)"
			}
			bail = false
		case turboNoEquivalent[name] != "":
			return "", nil, turboNoEquivalent[name]
		case turboQuiet[name]:
			dropped = append(dropped, "--"+name)
		default:
			return "", nil, "the turbo flag --" + name + " has no smith equivalent"
		}
	}
	if len(tasks) == 0 {
		return "", nil, "turbo with no task"
	}
	line := "smith run " + strings.Join(tasks, " ")
	if len(filters) > 0 {
		line += " -p " + strings.Join(filters, ",")
	}
	if bail {
		flags = append(flags, "--bail")
	}
	if len(flags) > 0 {
		line += " " + strings.Join(flags, " ")
	}
	return line, dropped, ""
}

// filterOf translates one --filter value. A package name (or its directory's
// name) becomes the smith project; a name glob and an exclusion stay as they
// are; Turborepo's graph and path filters (`web...`, `./apps/*`, `[main]`)
// have no smith counterpart.
func filterOf(value string, names map[string]string) (string, string) {
	negate := strings.HasPrefix(value, "!")
	value = strings.TrimPrefix(value, "!")
	prefix := ""
	if negate {
		prefix = "!"
	}
	switch {
	case value == "":
		return "", "an empty --filter"
	case strings.ContainsAny(value, "[]{}^") || strings.Contains(value, "...") || strings.HasPrefix(value, "."):
		return "", "--filter=" + value + " selects by dependency graph, path or git range, which smith run -p does not"
	case strings.Contains(value, "*"):
		if strings.Contains(value, "/") {
			return "", "--filter=" + value + " is a glob over package names with a scope; name the smith projects instead"
		}
		return prefix + value, ""
	}
	if name, ok := names[value]; ok {
		return prefix + name, ""
	}
	return "", "--filter=" + value + " names no package of the workspace"
}
