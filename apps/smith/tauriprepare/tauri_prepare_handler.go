package tauriprepare

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri"
	"github.com/spf13/cobra"
)

// Command is `smith tauri-prepare <project dir> [--verify]`.
func Command(version string) *cobra.Command {
	var verify bool
	cmd := &cobra.Command{
		Use:    "tauri-prepare <project-dir>",
		Short:  "Build a Tauri app's sidecar programs and copy them and its resources where tauri build looks, or check the built app holds them",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := workspace.FindRoot(".")
			if err != nil {
				return err
			}
			var settings map[string]any
			if raw := os.Getenv(tauri.SettingsEnv); raw != "" {
				if err := json.Unmarshal([]byte(raw), &settings); err != nil {
					return fmt.Errorf("%s is not JSON: %w", tauri.SettingsEnv, err)
				}
			}
			extras, err := tauri.ParseExtras(settings)
			if err != nil {
				return err
			}
			triple, err := tauri.HostTriple()
			if err != nil || triple == "" {
				return fmt.Errorf("cannot tell the target triple: is Rust installed (rustc -vV)? %v", err)
			}
			projectRoot := filepath.ToSlash(args[0])
			if verify {
				lines, err := tauri.VerifyExtras(root, projectRoot, extras)
				for _, l := range lines {
					fmt.Fprintln(cmd.OutOrStdout(), l)
				}
				return err
			}
			program, closeSession := programOf(root, version)
			defer closeSession()
			lines, err := tauri.PrepareExtras(tauri.PrepareRequest{WorkspaceRoot: root, ProjectRoot: projectRoot, Extras: extras, Triple: triple, Program: program})
			for _, l := range lines {
				fmt.Fprintln(cmd.OutOrStdout(), l)
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&verify, "verify", false, "check that the built app holds the declared sidecars instead of preparing them")
	return cmd
}

// programOf builds a sidecar's Rust project. The workspace is opened once, on
// the first sidecar that names a project.
func programOf(root, version string) (func(tauri.Sidecar) (string, error), func()) {
	var s *session.Session
	program := func(sc tauri.Sidecar) (string, error) {
		if s == nil {
			opened, err := session.Open(session.Options{Dir: root, Version: version, Log: io.Discard})
			if err != nil {
				return "", err
			}
			s = opened
		}
		p := s.Graph.Get(sc.Project)
		if p == nil {
			return "", fmt.Errorf("sidecar %s: the workspace has no project %s (smith projects lists them)", sc.Name, sc.Project)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(p.Root), "Cargo.toml")); err != nil {
			return "", fmt.Errorf("sidecar %s: %s is not a Rust project; smith builds only Rust programs as sidecars (give the file as from)", sc.Name, sc.Project)
		}
		crate := p.Name
		if len(p.Exports) > 0 {
			crate = p.Exports[0]
		}
		return tauri.BuildSidecar(root, crate, sc)
	}
	return program, func() {
		if s != nil {
			s.Close()
		}
	}
}
