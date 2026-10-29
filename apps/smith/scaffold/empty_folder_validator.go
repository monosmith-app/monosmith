package scaffold

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
)

// ignoredInEmptyFolder are the entries a folder may hold and still count as
// empty for smith new --here: the repository, the editor's settings and the
// files an operating system adds on its own.
var ignoredInEmptyFolder = map[string]bool{
	".git": true, ".vscode": true, ".DS_Store": true, "Thumbs.db": true, "desktop.ini": true,
}

// checkFolderEmpty refuses a folder that already holds files, so smith new
// --here never writes among someone's work.
func checkFolderEmpty(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !ignoredInEmptyFolder[e.Name()] {
			return problem.New(abs+" is not empty, and smith new --here only fills an empty folder",
				"smith new <name>     # creates the workspace in a new folder",
				"smith adopt          # brings the files in this folder into MonoSmith")
		}
	}
	return nil
}

// newArgs lets smith new --here take no name, which then comes from the folder.
func newArgs(inner cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if here, _ := cmd.Flags().GetBool("here"); here {
			return cobra.MaximumNArgs(1)(cmd, args)
		}
		return inner(cmd, args)
	}
}

// nameFromFolder names the workspace after the current folder when no name
// was given and the folder's name is one smith accepts; otherwise the name is
// asked for like any missing argument.
func nameFromFolder(args []string) []string {
	if len(args) > 0 {
		return args
	}
	cwd, err := filepath.Abs(".")
	if err != nil {
		return args
	}
	name := strings.ToLower(filepath.Base(cwd))
	if ValidateName("workspace", name) != nil {
		return args
	}
	return []string{name}
}
