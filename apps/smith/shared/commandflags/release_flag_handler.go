package commandflags

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/spf13/cobra"
)

// AddReleaseFlag registers --release: run each task's release profile, the
// optimised build (ADR 0092). A task without one runs as it is.
//
// dst receives the value; usage says what the profile changes for this
// command.
func AddReleaseFlag(cmd *cobra.Command, dst *bool, usage string) {
	c := flagvocabulary.Find("release")
	cmd.Flags().BoolVarP(dst, c.Flag, c.Short, false, usage+alsoNote(c))
	addSynonyms(cmd, c)
}
