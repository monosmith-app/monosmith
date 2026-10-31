package tauri

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// frontendDistFindings warns for each project whose tauri.conf.json names a
// frontendDist folder that is not where smith builds the web app
// (dist/<project dir>): `tauri build` would fail with "Unable to find your web
// assets". A URL, a missing value or a configuration that is not tauri.conf.json
// is left alone, as is a project with no build task.
func frontendDistFindings(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	var findings []protocol.Finding
	for _, p := range projects {
		if _, builds := p.Tasks["build"]; !builds || !tauriapp.Detect(workspaceRoot, p.Root) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), tauriapp.SrcDir, "tauri.conf.json"))
		if err != nil {
			continue
		}
		var conf struct {
			Build struct {
				FrontendDist any `json:"frontendDist"`
			} `json:"build"`
		}
		if json.Unmarshal(raw, &conf) != nil {
			continue
		}
		dist, ok := conf.Build.FrontendDist.(string)
		if !ok || strings.Contains(dist, "://") {
			continue
		}
		want := "../" + protocol.BuildDirFromProject(p.Root)
		findings = append(findings, protocol.Finding{
			Check:   "frontendDist of " + p.Name + " is the project's build output",
			OK:      path.Clean(path.Join(tauriapp.SrcDir, dist)) == path.Clean(path.Join(tauriapp.SrcDir, want)),
			Warning: true,
			Detail:  p.Root + "/" + tauriapp.SrcDir + "/tauri.conf.json has frontendDist " + dist + ", but smith builds the web app into " + protocol.ProjectBuildDir(p.Root),
			Remedy:  `set build.frontendDist to "` + want + `" in ` + p.Root + "/" + tauriapp.SrcDir + "/tauri.conf.json, and leave beforeBuildCommand empty (smith builds the web app first)",
		})
	}
	return findings
}
