package plugindraft

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

var langName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Command is `smith plugin draft`.
func Command() *cobra.Command {
	var lang, output string
	var stdout bool
	cmd := &cobra.Command{
		Use:   "draft <sample-folder> --lang <name>",
		Short: "Draft what a language plugin needs from a sample project: manifests, version line, tasks, publish step",
		Long: `Look at a small working project in a language smith does not know and write
<lang>.plugin-draft.toml: the files that name a project and give it a version,
the line that holds the version, the tasks it already runs (a Makefile or
justfile target, the scripts of a JSON manifest) and the lines of its CI that
publish. Every value says where it came from.

The draft is not a plugin. It is the list of facts a plugin author, or an
assistant writing the plugin, starts from. It prints what is still to write.
Then scaffold the plugin and run its conformance suite:

  smith add lib lang-<name> --stack go-smith-plugin
  smith plugin check ./lang-<name>`,
		Example: "  smith plugin draft ../samples/zig-hello --lang zig",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !langName.MatchString(lang) {
				return errors.New("--lang is required: a lower-case language name such as zig")
			}
			info, err := os.Stat(args[0])
			if err != nil || !info.IsDir() {
				return fmt.Errorf("%s is not a folder: give the folder of a small working project", args[0])
			}
			d := Inspect(os.DirFS(args[0]))
			text := Render(lang, args[0], d)
			out := cmd.OutOrStdout()
			if stdout {
				fmt.Fprint(out, text)
				return nil
			}
			if output == "" {
				output = lang + ".plugin-draft.toml"
			}
			if err := os.WriteFile(filepath.Clean(output), []byte(text), 0o644); err != nil {
				return err
			}
			fmt.Fprintf(out, "Wrote %s: %d manifests, %d tasks, %d publish lines found in %s.\n", output, len(d.Manifests), len(d.Tasks), len(d.Publish), args[0])
			fmt.Fprintln(out, "Still to write:")
			for _, t := range Todo(d) {
				fmt.Fprintf(out, "  - %s\n", t)
			}
			fmt.Fprintf(out, "→ smith add lib lang-%s --stack go-smith-plugin, then smith plugin check ./lang-%s --generate\n", lang, lang)
			return nil
		},
	}
	cmd.Flags().StringVar(&lang, "lang", "", "the language's name, lower case (required)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "the draft file (default <lang>.plugin-draft.toml)")
	cmd.Flags().BoolVar(&stdout, "stdout", false, "print the draft instead of writing a file")
	return cmd
}
