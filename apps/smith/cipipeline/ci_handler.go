package cipipeline

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// changeVerb says what Write did to a file it reports as changed: a file that
// is gone was removed (a provider dropped from [ci] providers), not written.
func changeVerb(root, path string) string {
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); errors.Is(err, fs.ErrNotExist) {
		return "removed"
	}
	return "wrote"
}

// Command is `smith ci`.
func Command(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ci",
		Short: "Run a CI phase, or regenerate the pipelines",
		Long: `The generated pipelines call these phases, so CI runs exactly what you can run locally:

  setup    install every toolchain dependency the workspace's plugins need
  audit    check dependencies for known vulnerabilities
  verify   lint, typecheck, test and build — only affected projects on a pull request
  pack     run every project's package task into dist/drop, kept as the pipeline's
           "drop" artifact on pushes to the release branch
  native   lint, test, build and package the projects tagged native (cgo) on this
           OS; the pipeline runs it on each OS and, on a release, attaches the build
  release  version, publish and tag every package with releasable commits`,
	}
	for _, phase := range Phases {
		cmd.AddCommand(&cobra.Command{
			Use:   phase,
			Short: "Run the " + phase + " phase",
			Args:  cobra.NoArgs,
			RunE: func(c *cobra.Command, _ []string) error {
				s, err := session.Open(session.Options{Dir: ".", Version: version, Log: c.ErrOrStderr()})
				if err != nil {
					return err
				}
				defer s.Close()
				ctx, stop := signal.NotifyContext(c.Context(), os.Interrupt)
				defer stop()
				return RunPhase(ctx, s, phase, c.OutOrStdout())
			},
		})
	}
	var check bool
	generate := &cobra.Command{
		Use:   "generate",
		Short: "Rewrite the CI pipelines from the workspace's projects (--check only reports drift)",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: c.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			if check {
				contributions, err := Collect(s)
				if err != nil {
					return err
				}
				if drift := Drift(s.Root, s.Config, NewModel(contributions, SmithVersionFor(s.Config, s.Version), s.Config.Release.Branch)); len(drift) > 0 {
					return errors.New("pipelines are out of date: " + strings.Join(drift, "; ") + " — run `smith ci generate`")
				}
				fmt.Fprintln(c.OutOrStdout(), "pipelines are up to date")
				return nil
			}
			contributions, err := Collect(s)
			if err != nil {
				return err
			}
			model := NewModel(contributions, SmithVersionFor(s.Config, s.Version), s.Config.Release.Branch)
			changed, lost, err := Write(s.Root, s.Config, model)
			if err != nil {
				return err
			}
			for _, limit := range Limits(s.Config, model) {
				fmt.Fprintln(c.ErrOrStderr(), "note: "+limit)
			}
			for _, f := range changed {
				fmt.Fprintln(c.OutOrStdout(), changeVerb(s.Root, f)+" "+f)
			}
			for _, l := range lost {
				fmt.Fprintln(c.ErrOrStderr(), "warning: slot no longer exists, its content was not kept: "+l)
			}
			if len(changed) == 0 {
				fmt.Fprintln(c.OutOrStdout(), "pipelines are up to date")
			}
			return nil
		},
	}
	generate.Flags().BoolVar(&check, "check", false, "report drift without writing")
	cmd.AddCommand(generate)
	cmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "List what to set up by hand for the pipelines: secrets, and each private registry's permissions",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: c.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			contributions, err := Collect(s)
			if err != nil {
				return err
			}
			model := NewModel(contributions, SmithVersionFor(s.Config, s.Version), s.Config.Release.Branch)
			steps := append(ManualSteps(s.Config, model), Limits(s.Config, model)...)
			if len(steps) == 0 {
				fmt.Fprintln(c.OutOrStdout(), "nothing to set up by hand")
				return nil
			}
			fmt.Fprintln(c.OutOrStdout(), "To set up by hand before the pipelines run:")
			for _, step := range steps {
				fmt.Fprintln(c.OutOrStdout(), "  - "+step)
			}
			return nil
		},
	})
	return cmd
}
