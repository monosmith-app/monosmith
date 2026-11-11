package tauri

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/extras"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// frontendDistOf is the folder the build.frontendDist of a project's
// src-tauri/tauri.conf.json names, workspace-relative (Tauri reads it relative
// to the configuration file). It is "" for a URL, a missing value or a
// configuration that is not tauri.conf.json.
func frontendDistOf(workspaceRoot, root string) string {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(root), tauriapp.SrcDir, "tauri.conf.json"))
	if err != nil {
		return ""
	}
	var conf struct {
		Build struct {
			FrontendDist any `json:"frontendDist"`
		} `json:"build"`
	}
	if json.Unmarshal(raw, &conf) != nil {
		return ""
	}
	dist, ok := conf.Build.FrontendDist.(string)
	if !ok || dist == "" || strings.Contains(dist, "://") {
		return ""
	}
	return path.Clean(path.Join(root, tauriapp.SrcDir, dist))
}

// frontendDistFindings warns for each project whose tauri.conf.json names a
// frontendDist folder that is not where smith builds the web app of a Tauri
// project (dist/tauri/<project dir>/frontend, ADR 0092): the frontend is a
// work file of the Tauri build, and the shipped folder holds the installers.
// A URL, a missing value or a configuration that is not tauri.conf.json is left
// alone, as is a project with no build task.
func frontendDistFindings(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	var findings []protocol.Finding
	for _, p := range projects {
		if _, builds := p.Tasks["build"]; !builds || !tauriapp.Detect(workspaceRoot, p.Root) {
			continue
		}
		dist := frontendDistOf(workspaceRoot, p.Root)
		if dist == "" {
			continue
		}
		folder := extras.FrontendDir(p.Root)
		want := "../" + protocol.FromProject(p.Root, folder)
		ok := dist == folder
		findings = append(findings, protocol.Finding{
			Check:   "frontendDist of " + p.Name + " is " + folder,
			Project: p.Name,
			OK:      ok,
			Warning: true,
			Problem: toolrun.ProblemUnless(ok, "frontendDist is "+dist+", not "+folder),
			Detail:  p.Root + "/" + tauriapp.SrcDir + "/tauri.conf.json takes the web app from " + dist + ". smith builds the web app of a Tauri app into " + folder + ", and keeps " + protocol.ProjectBuildDir(p.Root) + " for the installers",
			Remedy:  `set build.frontendDist to "` + want + `" in ` + p.Root + "/" + tauriapp.SrcDir + "/tauri.conf.json, build the web app into " + folder + " (outDir in vite.config), and leave beforeBuildCommand empty (smith builds the web app first)",
		})
	}
	return findings
}
