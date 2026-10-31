package tauri

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Task names of the sidecar and resource steps.
const (
	PrepareTask = "tauri-prepare"
	VerifyTask  = "tauri-verify"
)

// Sidecar is a program the app bundles (bundle.externalBin). It comes from
// a Rust project of the workspace (Project, built in release mode; Bin names
// the program when the crate builds several) or from a file that something
// else built (From).
type Sidecar struct {
	Name    string `json:"name"`
	Project string `json:"project,omitempty"`
	Bin     string `json:"bin,omitempty"`
	From    string `json:"from,omitempty"`
}

// Resource is a file, a folder (its contents) or a pattern with * and ? that
// the app bundles (bundle.resources), copied to To inside the app's resources.
type Resource struct {
	From string `json:"from"`
	To   string `json:"to,omitempty"`
}

// Extras are what [projects.<p>.settings.tauri] adds to the bundle.
type Extras struct {
	Sidecars  []Sidecar  `json:"sidecars,omitempty"`
	Resources []Resource `json:"resources,omitempty"`
}

// Empty says there is nothing to prepare.
func (e Extras) Empty() bool { return len(e.Sidecars) == 0 && len(e.Resources) == 0 }

// PrepareDir is where a project's prepared sidecars, resources and generated
// Tauri configuration go, workspace-relative: dist/tauri-prepare/<project dir>.
func PrepareDir(root string) string {
	return protocol.BuildDir + "/tauri-prepare/" + strings.Trim(root, "/")
}

// fromSrcTauri is the path from a project's src-tauri folder back to the
// workspace root, with a trailing slash ("../../../" for apps/studio).
func fromSrcTauri(root string) string {
	return strings.Repeat("../", strings.Count(strings.Trim(root, "/"), "/")+2)
}
