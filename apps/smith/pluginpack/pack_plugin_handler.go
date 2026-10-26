package pluginpack

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/pluginarchive"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
)

// Command is `smith plugin pack`. read reads the plugin.json in a directory
// (it returns an error wrapping fs.ErrNotExist when there is none); the
// manifest format belongs to the plugin host, not to the packer.
func Command(read func(dir string) (Plugin, error)) *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "pack [dir]",
		Short: "Pack a plugin directory into an archive `smith plugin install` accepts",
		Long: `Packs the plugin in dir (default: the current directory, which must hold a
plugin.json with a name and a version) into <name>-<version>.smith-plugin.zip,
and writes the archive's SHA-256 to <archive>.sha256 beside it.

The archive holds every file of the directory except .git, earlier archives,
checksums.sha256 and what the directory's .smithpluginignore lists (one pattern
per line: "name" at any depth, "dir/" for a directory, "a/b" or "/a" from the
root, * and ** as globs, # for comments). The program plugin.json names for
each platform is marked executable, and a checksums.sha256 lists the SHA-256 of
every file. Files are stored uncompressed in name order with a fixed timestamp,
so the same directory always gives the same archive, byte for byte. Symbolic
links are refused.

Build the program for each platform into the directory first. Nothing is run:
use "smith plugin check" for that.`,
		Example: "  smith plugin pack\n  smith plugin pack ./widget --out dist",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			installed, err := read(dir)
			if errors.Is(err, fs.ErrNotExist) {
				return problem.New(fmt.Sprintf("%s has no %s", dir, pluginarchive.ManifestFile), "run `smith plugin pack` in the plugin's directory, or give it: `smith plugin pack <dir>`")
			}
			if err != nil {
				return problem.New(err.Error())
			}
			if installed.Version == "" {
				return problem.New(fmt.Sprintf("%s has no version", filepath.Join(dir, pluginarchive.ManifestFile)), `add "version" to plugin.json: it names the archive and is what users see`)
			}
			res, err := Pack(dir, installed, out)
			if err != nil {
				return problem.New(err.Error())
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Packed %s %s: %d files, %d bytes.\n", installed.Name, installed.Version, len(res.Files), res.Bytes)
			fmt.Fprintf(w, "  archive  %s\n  sha256   %s  (also in %s)\n", res.Path, res.SHA256, filepath.Base(res.DigestPath))
			for _, p := range res.Programs {
				fmt.Fprintf(w, "  program  %s\n", p)
			}
			if len(res.Programs) == 0 {
				fmt.Fprintln(w, "  program  none packed: the command is looked up on PATH")
			}
			fmt.Fprintln(w, "The plugin was not started. Publish the archive and its .sha256 file together; users install it with `smith plugin install`.")
			return nil
		},
	}
	cmd.Flags().StringVarP(&out, "out", "o", ".", "directory to write the archive to")
	return cmd
}
