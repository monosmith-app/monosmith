package commandflags

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/spf13/cobra"
)

// AddFixFlag registers --fix: fix what the command can fix on its own, such
// as the problems a linter can rewrite.
//
// dst receives the value; usage says what the fix changes for this command.
func AddFixFlag(cmd *cobra.Command, dst *bool, usage string) {
	c := flagvocabulary.Find("fix")
	cmd.Flags().BoolVarP(dst, c.Flag, c.Short, false, usage+alsoNote(c))
	addSynonyms(cmd, c)
}
