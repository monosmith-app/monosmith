package sbom

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith release sbom`.
func Command(version string) *cobra.Command {
	var outDir, ver string
	cmd := &cobra.Command{
		Use:   "sbom [project...]",
		Short: "Write the CycloneDX SBOM of released apps, as a release attaches it",
		Long: `sbom writes <product>_<version>_sbom.cdx.json for each named project, or for
every app with a [release] table, into dist/release/<project>. Each project's
plugin lists its components from its lockfile or manifest. A release writes
and attaches the same file; this is that step on its own, to inspect or to
validate. The version is the one in the project's manifest unless --version
is given.`,
		RunE: func(cmd *cobra.Command, names []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			var projects []*projectgraph.Project
			if len(names) == 0 {
				for _, p := range s.Graph.Projects {
					if p.Release != nil {
						projects = append(projects, p)
					}
				}
			}
			for _, n := range names {
				p := s.Graph.Get(n)
				if p == nil {
					return fmt.Errorf("no project %q", n)
				}
				projects = append(projects, p)
			}
			out := cmd.OutOrStdout()
			for _, p := range projects {
				v := ver
				if v == "" {
					v = p.Version
				}
				if v == "" {
					v = "0.0.0"
				}
				product := ""
				if p.Release != nil {
					product = p.Release.Product
				}
				dir := outDir
				if dir == "" {
					dir = "dist/release/" + p.Name
				}
				rel, err := Write(s, p, Request{Version: v, Product: product, OutDir: dir, SmithVersion: version})
				if err != nil {
					return fmt.Errorf("%s: %w", p.Name, err)
				}
				if rel == "" {
					fmt.Fprintf(out, "%s: the %s plugin does not list components; no SBOM\n", p.Name, p.Plugin)
					continue
				}
				fmt.Fprintln(out, rel)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&ver, "version", "", "the version to name the SBOM for (default: the manifest's)")
	cmd.Flags().StringVar(&outDir, "out", "", "workspace-relative output directory (default: dist/release/<project>)")
	return cmd
}
