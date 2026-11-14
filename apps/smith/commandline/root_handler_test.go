package commandline

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
	"github.com/monosmith-app/monosmith/apps/smith/commandflags"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/spf13/cobra"
)

// visiblePaths are the command paths a user can run, without cobra's own.
func visiblePaths(c *cobra.Command, prefix []string) [][]string {
	var out [][]string
	for _, s := range c.Commands() {
		if s.Hidden || s.Name() == "help" || s.Name() == "completion" {
			continue
		}
		path := append(append([]string{}, prefix...), s.Name())
		out = append(out, path)
		out = append(out, visiblePaths(s, path)...)
	}
	return out
}

func TestEveryCommandIsAnActionAndEveryActionIsACommand(t *testing.T) {
	root := Root("0.0.0")
	exists := map[string]bool{}
	for _, path := range visiblePaths(root, nil) {
		exists[strings.Join(path, " ")] = true
		if _, ok := catalog.Find(path); !ok {
			t.Errorf("smith %s has no action: add one to libs/actions/catalog/catalog_config.go", strings.Join(path, " "))
		}
	}
	for _, a := range catalog.All() {
		if !exists[a.ID()] {
			t.Errorf("action %q names a command that is not there: remove or rename it in libs/actions/catalog/catalog_config.go", a.ID())
		}
	}
}

// A command that can change the workspace and takes --yes asks without a
// terminal, so the shell must append --yes after its own confirmation: that
// needs Asks on the action, or every run the shell starts fails with "nobody can
// confirm this change".
func TestEveryCommandThatTakesYesAndChangesTheWorkspaceIsMarkedAsAsking(t *testing.T) {
	root := Root("0.0.0")
	for _, path := range visiblePaths(root, nil) {
		cmd, _, err := root.Find(path)
		if err != nil || cmd == nil || cmd.Flags().Lookup("yes") == nil {
			continue
		}
		action, ok := catalog.Find(path)
		if !ok || action.Asks {
			continue
		}
		changes := action.Safety == catalog.Changes
		for _, v := range action.Variants {
			// A variant that already needs --yes is the confirmed form.
			changes = changes || (v.Safety == catalog.Changes && !slices.Contains(v.When, "--yes"))
		}
		if changes {
			t.Errorf("smith %s takes --yes and changes the workspace, but its action is not marked Asks: the shell would start it without --yes (libs/actions/catalog/catalog_config.go)", strings.Join(path, " "))
		}
	}
}

func TestEveryActionHasAFormOrAMenuEntryInTheShellOrIsExcused(t *testing.T) {
	shown := guide.PresentedWithForms(clicontract.DescribeCommands(Root("0.0.0")))
	for _, a := range catalog.All() {
		switch {
		case shown[a.ID()] && a.NotInTUI != "":
			t.Errorf("action %q is in the shell and also excused (%q): drop NotInTUI", a.ID(), a.NotInTUI)
		case !shown[a.ID()] && a.NotInTUI == "":
			t.Errorf("action %q has no menu entry or form in the shell, and a form could not be made for it: say why in NotInTUI (libs/actions/catalog/catalog_config.go)", a.ID())
		}
	}
}

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
	"cache push":         "only adds signed entries the remote cache lacks, as every run does; it changes no workspace file and deletes nothing; --dry-run shows it first",
	"ci":                 "with no phase it runs the setup phase: it installs the dependencies the lockfiles name, the same as the pipelines run it without a terminal, and changes no file git tracks",
	"link":               "changes only the smith entries the package manager put on PATH, never a workspace file: it keeps the original of each, names what it changed, and --undo gives them back; the owner decided it needs no consent because nothing is lost (the first run of a version does the same)",
	"docker description": "sets a remote description from the release pipeline, idempotent, with --dry-run; it changes no workspace file",
}

// A command that refuses without a terminal names --yes as the fix, so --help
// of every command that takes --yes must show it, with its -y shorthand.
func TestEveryCommandThatTakesYesShowsItInHelp(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	root := Root("0.0.0")
	for _, path := range visiblePaths(root, nil) {
		cmd, _, err := root.Find(path)
		if err != nil || cmd == nil {
			continue
		}
		f := cmd.Flags().Lookup("yes")
		if f == nil {
			continue
		}
		if f.Hidden || f.Shorthand != "y" || !strings.Contains(cmd.Flags().FlagUsages(), "-y, --yes") {
			t.Errorf("smith %s takes --yes but its help does not show -y, --yes: register it with commandflags.AddYesFlag", strings.Join(path, " "))
		}
	}
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

// On every command that takes a shared flag, --help shows the flag once, names
// its synonyms, and gives a synonym no help line of its own.
func TestSharedFlagsShowOnceInHelpAndNameTheirSynonyms(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	root := Root("0.0.0")
	for _, c := range flagvocabulary.All() {
		for _, path := range c.Commands {
			cmd, _, err := root.Find(strings.Fields(path))
			if err != nil || cmd == nil {
				t.Fatalf("smith %s: %v", path, err)
			}
			help := cmd.Flags().FlagUsages()
			if !strings.Contains(help, "--"+c.Flag) || (len(c.Synonyms) > 0 && !strings.Contains(help, "(also ")) {
				t.Errorf("smith %s: --help does not show --%s with its synonyms:\n%s", path, c.Flag, help)
			}
			for _, line := range strings.Split(help, "\n") {
				for _, s := range c.Synonyms {
					if strings.Contains(line, "--"+s.Name+" ") && !strings.Contains(line, "(also ") {
						t.Errorf("smith %s: synonym --%s has a help line of its own: %s", path, s.Name, line)
					}
				}
			}
		}
	}
}

// One concept, one flag (ADR 0084): no command defines another spelling of a
// shared concept, a shared flag has the vocabulary's shorthand and synonyms, and
// one shorthand never means two things. The tree is built in memory; no session
// opens and no plugin starts.
//
// A failure names the command and the flag. Fix it by registering the flag
// through package commandflags; if the second spelling is deliberate, add the
// flag to "exceptions" in libs/actions/flagvocabulary/flag_vocabulary.json with
// the reason.
func TestEveryCommandUsesTheFlagVocabulary(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	if got := commandflags.Violations(Root("0.0.0")); len(got) > 0 {
		t.Fatalf("flags break the vocabulary (flag_vocabulary.json, ADR 0084):\n  %s", strings.Join(got, "\n  "))
	}
}
