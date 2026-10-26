package scaffold

import (
	"fmt"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// partWords say, for the plugins people know by name, what each adds to a
// project.
var partWords = map[string]string{
	"eslint":     "linting",
	"typescript": "TypeScript",
	"jest":       "tests",
	"vitest":     "tests",
	"mocha":      "tests",
	"prettier":   "formatting",
}

// partStep is the line `smith add` prints while a plugin of the stack adds its
// part: what it adds in words, then the plugin that does it.
func partStep(plugin string) string {
	if words, ok := partWords[plugin]; ok {
		return fmt.Sprintf("Adding %s (%s)", words, plugin)
	}
	return "Adding " + plugin
}

// createStep is the first line `smith add` prints: what is made and where, then
// what it is made of in words (the stack's description, else the language).
func createStep(kind, name, dir string, st protocol.Stack, language string) string {
	made := language
	if st.Description != "" {
		made = st.Description
	}
	return fmt.Sprintf("Creating %s %s in %s: %s", kind, name, dir, made)
}

// doneStep is the last line of a successful add: what is ready and how many
// files it took.
func doneStep(kind, name string, files int, took string) string {
	return fmt.Sprintf("Done: %s %s is ready (%d files written, %s).", kind, name, files, took)
}
