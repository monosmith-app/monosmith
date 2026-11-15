package turboplan

import "strings"

// packageManagerCommands are the commands pnpm and yarn run themselves, as
// opposed to a package.json script of the same name; the value says what to
// use in a smith workspace.
var packageManagerCommands = map[string]string{
	"add": "smith install -w <project> <package>", "remove": "smith uninstall -w <project> <package>",
	"rm": "smith uninstall -w <project> <package>", "uninstall": "smith uninstall -w <project> <package>",
	"update": "smith up", "up": "smith up", "upgrade": "smith up", "upgrade-interactive": "smith up",
	"outdated": "smith up --check", "audit": "npm audit", "publish": "smith release", "pack": "npm pack",
	"link": "npm link", "unlink": "npm unlink", "list": "npm ls", "ls": "npm ls", "why": "npm explain",
	"dedupe": "npm dedupe", "prune": "npm prune", "rebuild": "npm rebuild", "config": "npm config",
	"init": "npm init", "create": "npm create", "patch": "", "patch-commit": "", "store": "", "env": "",
	"fetch": "", "import": "", "deploy": "", "set": "", "plugin": "", "constraints": "", "info": "",
	"npm": "", "cache": "", "bin": "", "root": "", "node": "", "explain": "", "version": "smith release",
}

// rewritePackageManager turns one pnpm or yarn command into smith or npm.
func rewritePackageManager(tool string, args []string, names map[string]string) (string, []string, string) {
	var filters []string
	recursive, ifPresent := false, false
	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		a := args[0]
		name, value, hasValue := strings.Cut(strings.TrimLeft(a, "-"), "=")
		args = args[1:]
		switch name {
		case "r", "recursive":
			recursive = true
		case "w", "workspace-root", "silent", "s", "reporter":
			// The root package is where npm run already runs.
			if name == "reporter" && !hasValue && len(args) > 0 {
				args = args[1:]
			}
		case "if-present":
			ifPresent = true
		case "filter", "F":
			if !hasValue {
				if len(args) == 0 {
					return "", nil, "--filter with no package"
				}
				value, args = args[0], args[1:]
			}
			v, problem := filterOf(value, names)
			if problem != "" {
				return "", nil, problem
			}
			filters = append(filters, v)
		default:
			return "", nil, tool + " " + a + " has no npm or smith equivalent convert can pick"
		}
	}
	if len(args) == 0 {
		if tool == "yarn" {
			return "smith install", nil, ""
		}
		return "", nil, "pnpm with no command"
	}
	command, rest := args[0], args[1:]
	if len(filters) > 0 || recursive {
		if command == "exec" {
			return "", nil, tool + " exec runs a command in every package, not a script; add a script to the packages and use smith run <task>"
		}
		if command == "run" {
			if len(rest) == 0 {
				return "", nil, tool + " " + command + " with no script"
			}
			command, rest = rest[0], rest[1:]
		}
		rewritten, problem := rewriteOneScript(tool, command, rest, ifPresent, filters, true)
		return rewritten, nil, problem
	}
	switch command {
	case "install", "i":
		frozen := false
		for _, r := range rest {
			if r != "--frozen-lockfile" && r != "--immutable" && r != "--ci" {
				return "", nil, tool + " install " + r + " has no npm or smith equivalent convert can pick"
			}
			frozen = true
		}
		if frozen {
			return "npm ci", nil, ""
		}
		return "smith install", nil, ""
	case "turbo":
		return rewriteTurbo(rest, names)
	case "exec", "dlx", "x":
		if len(rest) == 0 {
			return "", nil, tool + " " + command + " with no command"
		}
		if rest[0] == "turbo" {
			return rewriteTurbo(rest[1:], names)
		}
		return "npx " + strings.Join(rest, " "), nil, ""
	case "workspace":
		if tool != "yarn" || len(rest) < 2 {
			return "", nil, "yarn workspace with no package and command"
		}
		v, problem := filterOf(rest[0], names)
		if problem != "" {
			return "", nil, problem
		}
		script, tail := rest[1], rest[2:]
		if script == "run" && len(tail) > 0 {
			script, tail = tail[0], tail[1:]
		}
		rewritten, problem := rewriteOneScript(tool, script, tail, ifPresent, []string{v}, true)
		return rewritten, nil, problem
	case "workspaces":
		return "", nil, "yarn workspaces runs a command in every package; use smith run <task>"
	case "run", "run-script":
		if len(rest) == 0 {
			return "", nil, tool + " run with no script"
		}
		command, rest = rest[0], rest[1:]
	default:
		if instead, builtin := packageManagerCommands[command]; builtin {
			if instead == "" {
				return "", nil, tool + " " + command + " has no npm or smith equivalent"
			}
			return "", nil, tool + " " + command + " manages the packages of a " + tool + " workspace; use " + instead
		}
	}
	rewritten, problem := rewriteOneScript(tool, command, rest, ifPresent, nil, false)
	return rewritten, nil, problem
}

// rewriteOneScript runs a package.json script: `npm run` at the workspace
// root, `smith run` for a filter or a recursive run, which cannot pass
// arguments on.
func rewriteOneScript(tool, script string, args []string, ifPresent bool, filters []string, many bool) (string, string) {
	if many {
		if len(args) > 0 {
			return "", tool + " passes " + strings.Join(args, " ") + " to the script, and smith run does not pass arguments on; put them in the package.json script"
		}
		line := "smith run " + script
		if len(filters) > 0 {
			line += " -p " + strings.Join(filters, ",")
		}
		return line, ""
	}
	line := "npm run " + script
	if ifPresent {
		line += " --if-present"
	}
	if len(args) > 0 {
		line += " -- " + strings.Join(args, " ")
	}
	return line, ""
}
