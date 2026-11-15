package convertplan

import (
	"fmt"
	"strings"
)

// RewriteScript turns a root package.json script that drives Nx (or the old
// smith CLI) into the smith command doing the same. names maps Nx project
// names to smith names. ok is false when part of it has no smith equivalent.
func RewriteScript(script string, names map[string]string) (string, bool) {
	parts := strings.Split(script, "&&")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		words := strings.Fields(part)
		if len(words) > 0 && words[0] == "npx" {
			words = words[1:]
		}
		if len(words) == 0 {
			return "", false
		}
		if words[0] != "nx" && words[0] != "mnci" {
			out = append(out, strings.TrimSpace(part))
			continue
		}
		rewritten, ok := rewriteOne(words, names)
		if !ok {
			return "", false
		}
		out = append(out, rewritten)
	}
	return strings.Join(out, " && "), true
}

func rewriteOne(words []string, names map[string]string) (string, bool) {
	tool, args := words[0], words[1:]
	if len(args) == 0 {
		return "", false
	}
	if tool == "mnci" {
		// The old CLI's `mnci lint [--affected] [--base x]`.
		if args[0] == "lint" {
			return "smith run lint" + flags(args[1:]), true
		}
		return "", false
	}
	switch args[0] {
	case "run-many", "affected":
		targets, rest, ok := targetsOf(args[1:])
		if !ok {
			return "", false
		}
		cmd := "smith run " + strings.Join(targets, " ")
		if args[0] == "affected" {
			cmd += " --affected"
		}
		for _, r := range rest {
			name, value, _ := strings.Cut(r, "=")
			switch name {
			case "--projects", "-p":
				var projects []string
				for _, p := range strings.Split(value, ",") {
					if n, ok := smithName(p, names); ok {
						projects = append(projects, n)
					} else {
						return "", false
					}
				}
				cmd += " -p " + strings.Join(projects, ",")
			case "--base":
				cmd += " --base " + value
			case "--parallel", "--output-style", "--skip-nx-cache", "--nxBail", "--verbose":
			default:
				return "", false
			}
		}
		return cmd, true
	case "run":
		if len(args) != 2 {
			return "", false
		}
		project, target, ok := strings.Cut(args[1], ":")
		target, _, _ = strings.Cut(target, ":")
		if n, known := smithName(project, names); ok && known {
			return fmt.Sprintf("smith run %s -p %s", target, n), true
		}
		return "", false
	case "graph":
		return "smith graph", true
	case "show":
		if len(args) > 1 && args[1] == "projects" {
			return "smith projects", true
		}
	case "release":
		for _, a := range args[1:] {
			if a != "--dry-run" && a != "-d" {
				return "", false
			}
		}
		return "smith release" + flags(args[1:]), true
	}
	if len(args) == 2 {
		// `nx build api` is `nx run api:build`.
		if n, known := smithName(args[1], names); known {
			return fmt.Sprintf("smith run %s -p %s", args[0], n), true
		}
	}
	return "", false
}

// smithName finds the smith name of the project a script names the way Nx
// resolves it: by its Nx name, or by that name without its npm scope (core
// for @acme/core) when exactly one project has it.
func smithName(project string, names map[string]string) (string, bool) {
	if n, ok := names[project]; ok {
		return n, true
	}
	found, matches := "", 0
	for nx, n := range names {
		if scope, bare, scoped := strings.Cut(nx, "/"); scoped && strings.HasPrefix(scope, "@") && bare == project {
			found = n
			matches++
		}
	}
	return found, matches == 1
}

// ScriptGap says why a root script RewriteScript could not translate has no
// smith equivalent, when the reason is more precise than "cannot translate".
func ScriptGap(script string) string {
	for _, part := range strings.Split(script, "&&") {
		words := strings.Fields(part)
		if len(words) > 0 && words[0] == "npx" {
			words = words[1:]
		}
		if len(words) > 3 && words[0] == "nx" && words[1] == "run" && words[3] == "--" {
			return "it passes arguments to the task (" + strings.Join(words[3:], " ") + "), and smith run does not pass arguments on; add a script with them to the project and run that task instead"
		}
	}
	return ""
}

// targetsOf reads -t/--targets/--target values (comma or space separated)
// and returns the remaining flags.
func targetsOf(args []string) (targets, rest []string, ok bool) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		name, value, hasValue := strings.Cut(a, "=")
		switch name {
		case "-t", "--targets", "--target":
			if !hasValue {
				for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					i++
					targets = append(targets, strings.Split(args[i], ",")...)
				}
				continue
			}
			targets = append(targets, strings.Split(value, ",")...)
		case "-p", "--projects":
			if hasValue {
				rest = append(rest, a)
				continue
			}
			// Nx takes the projects comma or space separated: -p core cli.
			var projects []string
			for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				projects = append(projects, args[i])
			}
			rest = append(rest, name+"="+strings.Join(projects, ","))
		default:
			if strings.HasPrefix(a, "-") && !hasValue && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				a += "=" + args[i+1]
				i++
			}
			rest = append(rest, a)
		}
	}
	return targets, rest, len(targets) > 0
}

func flags(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return " " + strings.Join(args, " ")
}
