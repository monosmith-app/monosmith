package commandflags

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"github.com/spf13/cobra"
)

// AddProjectFlag registers -p, --project: the project or projects a command acts
// on. It is repeatable and comma separated, so -p web,api and -p web -p api mean
// the same. The synonyms the vocabulary lists (-w, --workspace, --projects) have
// the same behaviour; the help names them. In smith "workspace" is the whole
// repository, so -p is the form every printed command line uses.
//
// dst receives the values in the order given. usage says what the command
// accepts (names, globs, selectors, directories).
func AddProjectFlag(cmd *cobra.Command, dst *[]string, usage string) {
	c := flagvocabulary.Find("project")
	cmd.Flags().StringSliceVarP(dst, c.Flag, c.Short, nil, usage+alsoNote(c))
	addSynonyms(cmd, c)
}

// AddOneProjectFlag registers -p, --project for a command that acts on exactly
// one project (the app a deploy pipeline builds, the app a release publishes).
// The name, shorthand and synonyms are the vocabulary's; the value is one
// project.
//
// dst receives the project. usage says what the project is used for.
func AddOneProjectFlag(cmd *cobra.Command, dst *string, usage string) {
	c := flagvocabulary.Find("project")
	cmd.Flags().StringVarP(dst, c.Flag, c.Short, "", usage+alsoNote(c))
	addSynonyms(cmd, c)
}
