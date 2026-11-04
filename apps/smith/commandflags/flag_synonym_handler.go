package commandflags

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// markingValue is the value of a synonym: it writes through to the canonical
// flag's value and marks the canonical flag as set, so cmd.Flags().Changed with
// the canonical name is true whichever spelling the person typed.
type markingValue struct {
	pflag.Value
	base *pflag.Flag
}

// Set stores s in the shared value and marks the canonical flag as changed.
func (v markingValue) Set(s string) error {
	if err := v.Value.Set(s); err != nil {
		return err
	}
	v.base.Changed = true
	return nil
}

// addSynonyms registers each synonym of the concept as a flag that shares the
// value of the canonical flag, so mixing spellings on one command line adds up
// exactly as repeating the canonical one does. A synonym is a hidden flag: --help
// shows the canonical flag once, and its usage names the synonyms (alsoNote).
//
// It panics when cmd has no canonical flag for the concept: that is a mistake in
// the code that builds the command, found by any test that builds the tree.
func addSynonyms(cmd *cobra.Command, c flagvocabulary.Concept) {
	flags := cmd.Flags()
	base := flags.Lookup(c.Flag)
	if base == nil {
		panic(fmt.Sprintf("commandflags: %s has no flag --%s to give synonyms", cmd.CommandPath(), c.Flag))
	}
	for _, s := range c.Synonyms {
		flags.VarPF(markingValue{Value: base.Value, base: base}, s.Name, s.Short, base.Usage)
		synonym := flags.Lookup(s.Name)
		synonym.NoOptDefVal = base.NoOptDefVal
		synonym.Hidden = true
	}
}

// alsoNote words the synonyms of a concept for the end of a help line:
// " (also -w, --workspace, --projects)". It is empty when there are none.
func alsoNote(c flagvocabulary.Concept) string {
	if len(c.Synonyms) == 0 {
		return ""
	}
	var parts []string
	for _, s := range c.Synonyms {
		if s.Short != "" {
			parts = append(parts, "-"+s.Short)
		}
		parts = append(parts, "--"+s.Name)
	}
	return " (also " + strings.Join(parts, ", ") + ")"
}
