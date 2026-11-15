package npmpack

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
	"github.com/spf13/cobra"
)

// Command is `smith npm-pack <project dir>`.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:    "npm-pack <project-dir>",
		Short:  "Pack a Node project's build (dist/<project dir>) with npm, its package.json staged beside it (a VS Code extension with vsce)",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			staged, err := releasing.Stage(root, filepath.ToSlash(args[0]))
			if err != nil {
				return err
			}
			drop := filepath.Join(root, filepath.FromSlash(protocol.BuildDir), "drop")
			if err := os.MkdirAll(drop, 0o755); err != nil {
				return err
			}
			if _, ok := releasing.Extension(staged); ok {
				_, err := releasing.PackExtension(staged, drop)
				return err
			}
			return toolrun.Run(root, []string{"npm", "pack", staged, "--pack-destination", drop})
		},
	}
}
