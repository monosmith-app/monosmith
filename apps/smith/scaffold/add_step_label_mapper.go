package scaffold

import (
	"fmt"

	"github.com/monosmith-app/monosmith/libs/words/words"
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
	"docker":     "a container image",
}

// partStep is the line `smith add` prints after a plugin of the stack or a
// convention added its part: what it added in words, then the plugin. A plugin
// that changed nothing (no file, no install) says it had nothing to add for the
// kind, so a lib never reads as if docker added something to it.
func partStep(plugin, kind string, added protocol.GenerateResult) string {
	if len(added.Files) == 0 && len(added.Notes) == 0 && !added.InstallPending {
		return words.Say("addproject.part.nothing", "plugin", plugin, "kind", kind)
	}
	if what, ok := partWords[plugin]; ok {
		return words.Say("addproject.part.added_named", "what", what, "plugin", plugin)
	}
	return words.Say("addproject.part.added", "plugin", plugin)
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

// directoryReason is the line `smith add` prints when a Node project goes to
// another folder than its kind's own: the root package.json workspaces do not
// reach the kind's folder, so npm would not link the project there. It is empty
// when the project goes to the kind's folder.
func directoryReason(kind, conventional, parent string) string {
	if parent == conventional {
		return ""
	}
	return fmt.Sprintf("The %s goes to %s/, not %s/: the workspaces of the root package.json list %s/* and not %s/*, so npm links it there.", kind, parent, conventional, parent, conventional)
}

// doneStep is the last line of a successful add: what is ready and how many
// files it took.
func doneStep(kind, name string, files int, took string) string {
	return fmt.Sprintf("Done: %s %s is ready (%d files written, %s).", kind, name, files, took)
}
