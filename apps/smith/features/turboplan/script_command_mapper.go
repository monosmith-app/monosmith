package turboplan

import (
	"strings"
)

// Rewrite is a package.json script after its turbo, pnpm and yarn commands
// became smith and npm ones.
type Rewrite struct {
	// Script is the new command line.
	Script string
	// Changed is false when the script mentions none of the three.
	Changed bool
	// Problem says why part of the script could not be translated; Script is
	// then the original.
	Problem string
	// Dropped names turbo flags that only tune turbo's own output and caching
	// and that smith has no use for.
	Dropped []string
}

// RewriteScript translates the turbo, pnpm and yarn commands of a script into
// the smith and npm ones that do the same. names maps package names and
// directory names to smith project names. A script is split at `&&`; a part
// that needs another shell construct (a pipe, `;`, `||`) is left alone when it
// mentions none of the three tools, and reported when it does.
func RewriteScript(script string, names map[string]string) Rewrite {
	parts := strings.Split(script, "&&")
	out := make([]string, 0, len(parts))
	res := Rewrite{}
	for _, part := range parts {
		words := strings.Fields(part)
		if !mentionsTool(words) {
			out = append(out, strings.TrimSpace(part))
			continue
		}
		res.Changed = true
		if strings.ContainsAny(part, "|;`") || strings.Contains(part, "$(") {
			return Rewrite{Script: script, Changed: true, Problem: "a shell construct (pipe, ;, ||, substitution) around " + tools(words) + " is beyond what convert translates"}
		}
		line, dropped, problem := rewriteCommand(words, names)
		if problem != "" {
			return Rewrite{Script: script, Changed: true, Problem: problem}
		}
		res.Dropped = append(res.Dropped, dropped...)
		out = append(out, line)
	}
	res.Script = strings.Join(out, " && ")
	return res
}

func isTool(word string) bool { return word == "turbo" || word == "pnpm" || word == "yarn" }

func mentionsTool(words []string) bool {
	for _, w := range words {
		if isTool(w) {
			return true
		}
	}
	return false
}

func tools(words []string) string {
	var found []string
	for _, w := range words {
		if isTool(w) {
			found = append(found, w)
		}
	}
	return strings.Join(found, ", ")
}

// rewriteCommand translates one command (no `&&`).
func rewriteCommand(words []string, names map[string]string) (string, []string, string) {
	var env []string
	for len(words) > 0 && strings.Contains(words[0], "=") && !strings.HasPrefix(words[0], "-") {
		env = append(env, words[0])
		words = words[1:]
	}
	prefix := ""
	if len(env) > 0 {
		prefix = strings.Join(env, " ") + " "
	}
	if len(words) == 0 {
		return "", nil, "an empty command"
	}
	if words[0] == "npx" || words[0] == "bunx" {
		words = words[1:]
		if len(words) == 0 || words[0] != "turbo" {
			return "", nil, "a command npx runs besides turbo"
		}
	}
	switch words[0] {
	case "turbo":
		line, dropped, problem := rewriteTurbo(words[1:], names)
		return prefix + line, dropped, problem
	case "pnpm", "yarn":
		line, dropped, problem := rewritePackageManager(words[0], words[1:], names)
		return prefix + line, dropped, problem
	}
	return "", nil, "a command that runs " + words[0] + " before " + tools(words) + " (cross-env, dotenv, …); put smith run first and set the variables on the task"
}
