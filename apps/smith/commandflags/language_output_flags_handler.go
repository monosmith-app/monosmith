package commandflags

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/spf13/cobra"
)

// AddLangFlag registers -l, --lang: the language of a project or a workspace
// (synonym --language, as the vocabulary says).
//
// dst receives the value; def is its default; usage says what the language is
// used for.
func AddLangFlag(cmd *cobra.Command, dst *string, def, usage string) {
	c := flagvocabulary.Find("lang")
	cmd.Flags().StringVarP(dst, c.Flag, c.Short, def, usage+alsoNote(c))
	addSynonyms(cmd, c)
}

// AddOutputFlag registers -o, --output: the file or folder a command writes
// (synonym --out).
//
// dst receives the path; def is its default; usage says what is written.
func AddOutputFlag(cmd *cobra.Command, dst *string, def, usage string) {
	c := flagvocabulary.Find("output")
	cmd.Flags().StringVarP(dst, c.Flag, c.Short, def, usage+alsoNote(c))
	addSynonyms(cmd, c)
}

// AddYesFlag registers -y, --yes: do not ask for confirmation.
//
// dst receives the answer; usage says what taking the default means for this
// command, and what happens without a terminal.
func AddYesFlag(cmd *cobra.Command, dst *bool, usage string) {
	c := flagvocabulary.Find("yes")
	cmd.Flags().BoolVarP(dst, c.Flag, c.Short, false, usage)
}

// AddNoInstallFlag registers --no-install: change the manifests but do not run
// the package manager (synonym --skip-install).
//
// dst receives the value; usage says which install is left out.
func AddNoInstallFlag(cmd *cobra.Command, dst *bool, usage string) {
	c := flagvocabulary.Find("no-install")
	cmd.Flags().BoolVarP(dst, c.Flag, c.Short, false, usage+alsoNote(c))
	addSynonyms(cmd, c)
}
