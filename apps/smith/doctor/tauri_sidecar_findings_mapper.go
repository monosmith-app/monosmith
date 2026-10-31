package doctor

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/extras"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// tauriSidecarChecks fails for a Tauri app's sidecar that names a project the
// workspace does not have, or one that is not a Rust project: only the host
// sees every project, the tauri plugin's doctor sees its own.
func tauriSidecarChecks(s *session.Session) []protocol.Finding {
	var out []protocol.Finding
	for _, p := range s.Graph.Projects {
		declared, err := extras.ParseExtras(p.Settings[tauri.Name])
		if err != nil {
			continue // the plugin reports the invalid table
		}
		for _, sc := range declared.Sidecars {
			if sc.Project == "" {
				continue
			}
			f := protocol.Finding{Check: "the sidecar " + sc.Name + " of " + p.Name + " is a project that exists", OK: true}
			if sidecar := s.Graph.Get(sc.Project); sidecar == nil {
				f.OK, f.Detail, f.Remedy = false, "the workspace has no project "+sc.Project, "name one of `smith projects` in [projects."+p.Name+".settings.tauri] sidecars"
			} else if _, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(sidecar.Root), "Cargo.toml")); err != nil {
				f.OK, f.Detail, f.Remedy = false, sc.Project+" is not a Rust project (no Cargo.toml in "+sidecar.Root+"); smith builds only Rust programs as sidecars", "build the program another way and give its file as from"
			}
			out = append(out, f)
		}
	}
	return out
}
