package commandline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/commandflags"
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
)

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
