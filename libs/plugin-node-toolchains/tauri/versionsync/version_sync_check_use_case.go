package versionsync

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Findings warns for each Tauri app whose package.json, tauri.conf.json
// and src-tauri/Cargo.toml do not carry the same version. smith release writes
// all three, so the warning matters for what is built before one (tauri dev,
// tauri-build). A tauri.conf.json whose version is a path to package.json
// counts as package.json's own.
func Findings(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	var findings []protocol.Finding
	for _, p := range projects {
		if !tauriapp.Detect(workspaceRoot, p.Root) {
			continue
		}
		found := map[string]string{}
		if p.Version != "" {
			found["package.json"] = p.Version
		}
		if src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), tauriapp.SrcDir, "tauri.conf.json")); err == nil {
			if _, _, v, ok := confVersion(string(src)); ok && !isVersionPath(v) {
				found[tauriapp.SrcDir+"/tauri.conf.json"] = v
			}
		}
		if src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), tauriapp.SrcDir, "Cargo.toml")); err == nil {
			if v, ok := cargoVersion(string(src)); ok {
				found[tauriapp.SrcDir+"/Cargo.toml"] = v
			}
		}
		if len(found) < 2 {
			continue
		}
		versions := map[string]bool{}
		var files []string
		for file, v := range found {
			versions[v] = true
			files = append(files, file+" "+v)
		}
		sort.Strings(files)
		findings = append(findings, protocol.Finding{
			Check:   "the versions of " + p.Name + " agree",
			OK:      len(versions) <= 1,
			Warning: true,
			Detail:  p.Root + " has " + strings.Join(files, ", ") + "; smith release writes one version into all of them",
			Remedy:  `set "version": "../package.json" in ` + p.Root + "/" + tauriapp.SrcDir + "/tauri.conf.json and the same version as package.json in " + p.Root + "/" + tauriapp.SrcDir + "/Cargo.toml, or let smith release write them",
		})
	}
	return findings
}
