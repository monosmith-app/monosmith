package commandline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/actions/catalog"
)

// answersWithoutAsking are the commands that change something and still run
// without a terminal and without --yes, each with the reason. The rule of the
// owner: a command that changes the workspace asks in a terminal and refuses,
// naming --yes, anywhere else. A command is excused only when it fits one of
// these reasons, and a new command is not excused until it is written here.
var answersWithoutAsking = map[string]string{
	"config set":         "one named setting of smith.toml: the command line states key and value, the result line shows it, git takes it back",
	"config unset":       "one named setting of smith.toml: the command line states the key, the result line shows it, git takes it back",
	"architecture":       "one setting of smith.toml (the code layout): the command line names it and prints it",
	"plugin set":         "one named setting of one plugin in smith.toml: the command line states key and value",
	"plugin enable":      "one named setting of smith.toml: the command line names the plugin",
	"plugin disable":     "one named setting of smith.toml: the command line names the plugin",
	"plugin install":     "writes only the plugin folder of the person's machine, from the one source named, checked against its digest, and never starts the plugin",
	"plugin draft":       "writes one skeleton at the path named and keeps a file that exists unless --force",
	"cache prune":        "deletes only cache entries smith can rebuild, within the limit, prints what it freed; --dry-run shows it first",
	"docker description": "sets a remote description from the release pipeline, idempotent, with --dry-run; it changes no workspace file",
}

// An action that changes the workspace must take --yes and be marked Asks, so
// that it refuses without a terminal and the shell confirms for the person;
// the exceptions are named above with their reason.
func TestEveryCommandThatChangesTheWorkspaceAsksOrIsExcusedWithAReason(t *testing.T) {
	root := Root("0.0.0")
	for _, path := range visiblePaths(root, nil) {
		cmd, _, err := root.Find(path)
		if err != nil || cmd == nil || (cmd.Run == nil && cmd.RunE == nil) {
			continue
		}
		id := strings.Join(path, " ")
		action, ok := catalog.Find(path)
		if !ok || action.ID() != id {
			// A subcommand the catalogue lists under its parent is judged as the parent.
			continue
		}
		changes := action.Safety == catalog.Changes
		reason, excused := answersWithoutAsking[id]
		switch {
		case !changes && excused:
			t.Errorf("smith %s is excused (%s) but does not change the workspace: remove it from answersWithoutAsking", id, reason)
		case !changes:
		case excused && (action.Asks || cmd.Flags().Lookup("yes") != nil):
			t.Errorf("smith %s asks, so it needs no excuse: remove it from answersWithoutAsking", id)
		case excused:
		case !action.Asks:
			t.Errorf("smith %s changes the workspace but is not marked Asks: it would run without a terminal and without --yes. Make it ask (operationplan.Gate) and mark it Asks in libs/actions/catalog/catalog_config.go, or excuse it with a reason in answersWithoutAsking", id)
		case cmd.Flags().Lookup("yes") == nil:
			t.Errorf("smith %s is marked Asks but has no --yes flag, so nobody can answer for it", id)
		}
	}
}
