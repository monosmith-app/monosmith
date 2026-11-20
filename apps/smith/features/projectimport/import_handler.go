package projectimport

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/operation"
	"github.com/monosmith-app/monosmith/apps/smith/features/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/shared/commandargs"
	"github.com/monosmith-app/monosmith/apps/smith/shared/commandflags"
	"github.com/spf13/cobra"
)

// Command is `smith import <git-url|folder>`.
//
// Parameters: version is smith's.
// Returns: the command, which commits its work only with --commit.
func Command(version string) *cobra.Command {
	var o Options
	var imported string
	cmd := &cobra.Command{
		Use:   "import <git-url|folder>",
		Short: "Bring an existing project into this workspace, from a git repository or a folder",
		Long: `Bring an existing project into the smith workspace you are in. The source is a
git URL (cloned into a temporary folder) or a local folder. smith finds the
project by its manifest at the top of the source (package.json, pyproject.toml,
a .csproj, and every other manifest the language plugins know), then:

  place    a git repository keeps its history: its commits are written again
           under <folder> and merged, so git log -- <folder> and git blame
           follow them (--history squash makes one commit with git subtree
           add --squash, --history copy copies the files only). A plain
           folder is copied, without .git, installed dependencies and caches
  split    the development tools go to the workspace, README and LICENSE stay
           with the project, and its own lockfile is dropped (the
           versions an npm lockfile locked go into the workspace lockfile)
  join     the project joins its language's workspace (npm workspaces, the uv
           members, the .NET solution)
  tooling  a Node project keeps its own build and lint tools, or uses smith's
           (smith asks, and recommends its own for a legacy tool: ESLint 8,
           create-react-app, a tsc build into lib/; --tooling smith|keep
           answers). smith's go in first, then the old tools go
  check    the pipelines are written again, then smith install and
           smith doctor; it ends with what is left to do

A source whose manifest is below its top (src/MyLib/MyLib.csproj) is offered
the project there: --path <sub-folder> says it, and a git source keeps the
commits of that folder (git subtree split). A source with several projects is
a workspace: smith adopt takes it whole. A folder of sources with no manifest
(loose .ts, .py or .cs files) is offered the manifest of its language's
default stack, written from smith add's templates, and the sources stay where
they are: --manifest <language> says it, --manifest none refuses.

The folder is <kind folder>/<name> (apps/, libs/ or packages/). In a terminal
smith asks for it, with the kind the project looks like as the default.
Without a terminal, --as or --into says it: smith never picks it.

The import of a git history makes commits of its own (a merge commit) and
needs a clean work tree. smith commits nothing else unless you add --commit. A plan is shown first, and the
import runs after "y" in a terminal, or with --yes.`,
		Example: "  smith import https://github.com/owner/api --as app\n" +
			"  smith import ../color-tools --into libs/color-tools --dry-run\n" +
			"  smith import https://github.com/owner/web --as app --history squash --yes\n" +
			"  smith import https://github.com/owner/dotnet-lib --path src/MyLib --as package\n" +
			"  smith import ../scripts --as app --manifest python",
		Args: commandargs.Exactly(commandargs.Spec{Names: []string{"a git URL or a folder"}, Example: "smith import https://github.com/owner/api --as app"}),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.Source, o.Version = args[0], version
			o.Env = operationplan.DetectEnvironment()
			o.In, o.Out, o.Log = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
			p, err := Import(o)
			imported = p.Name
			return err
		},
	}
	cmd.Flags().StringVar(&o.As, "as", "", "the project's kind: app, lib or package (default: what the project looks like, shown in the plan)")
	cmd.Flags().StringVar(&o.Into, "into", "", "the project's folder, from the workspace root (default: <kind folder>/<name>, asked in a terminal)")
	cmd.Flags().StringVar(&o.Name, "name", "", "the project's name (default: the name in its manifest)")
	cmd.Flags().StringVar(&o.Ref, "ref", "", "the branch, tag or commit of a git source to import (default: its default branch)")
	cmd.Flags().StringVar(&o.History, "history", "", "keep (every commit, under the folder), squash (one commit) or copy (the files only); default keep for a git source, copy for a plain folder")
	cmd.Flags().StringVar(&o.ESLint, "eslint-setup", "", "what happens to a Node project's own ESLint setup: shared (the workspace's replaces it) or keep (default without a terminal)")
	cmd.Flags().StringSliceVar(&o.Tooling, "tooling", nil, "a Node project's build and lint tools: smith (smith's build and ESLint config, then the old tools go) or keep (default with --yes)")
	cmd.Flags().StringSliceVar(&o.Hoist, "hoist", nil, "development dependencies to move to the workspace without asking (the known tools move anyway)")
	cmd.Flags().StringVar(&o.Manifest, "manifest", "", "a source with no manifest: the language whose default stack's manifest smith writes (node, python, dotnet), or none; asked in a terminal")
	cmd.Flags().StringVar(&o.Path, "path", "", "the sub-folder of the source that holds the project (src/MyLib); a git source keeps the commits of that folder")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "show the plan and change nothing")
	commandflags.AddYesFlag(cmd, &o.Yes, "do not ask: apply the plan (--as or --into must say where the project goes)")
	return operation.Wrap(cmd, operation.Spec{Describe: func(*cobra.Command, []string) string { return "import " + imported }})
}
