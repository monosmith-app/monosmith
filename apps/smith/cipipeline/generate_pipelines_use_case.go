package cipipeline

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachecredentials"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/splitrepo"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Collect asks every plugin what CI needs for its projects.
func Collect(s *session.Session) (Contributions, error) {
	var c Contributions
	for _, client := range s.Plugins.Clients {
		if !client.Info.Capabilities.CI {
			continue
		}
		var projects []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.ServedBy(client.Info.Plugin.Name) {
				projects = append(projects, p.Project)
			}
		}
		if len(projects) == 0 {
			continue
		}
		var r protocol.CIResult
		if err := client.Call(protocol.MethodCI, protocol.CIParams{Projects: projects, Audit: auditSettings(s.Config.CI)}, &r, time.Minute); err != nil {
			return c, fmt.Errorf("plugin %s: %w", client.Info.Plugin.Name, err)
		}
		c.Add(r)
	}
	for _, p := range s.Graph.Projects {
		if _, ok := p.Tasks[PackTask]; ok {
			c.Pack = true
		}
		if release.Native(p) {
			c.Native = true
			c.NativePublish = c.NativePublish || p.Kind == protocol.ProjectPackage || (p.Kind == protocol.ProjectApp && len(p.Deployers) > 0)
			continue
		}
		c.Add(protocol.CIResult{Credentials: release.DistributionCredentials(p.Release)})
		if p.SplitRepo != "" {
			c.Add(protocol.CIResult{Credentials: []protocol.Credential{splitrepo.Credential(p.SplitRepo)}})
		}
		if p.Release != nil && len(p.Release.NPM) > 0 {
			c.Provenance = true
		}
		if p.Release != nil && s.Config.Release.GitHubReleases {
			c.Attest = true
		}
		if release.PubPackage(p) != "" {
			c.Pub = true
		}
	}
	if s.Config.Release.RubyGemsTrustedPublishing {
		c.UseRubyGemsTrustedPublishing()
	}
	if s.Config.Release.VSCEEntra {
		c.UseVSCEEntra()
	}
	// A remote cache is read by every step that runs tasks.
	c.Add(protocol.CIResult{Credentials: cachecredentials.For(s.Config.Cache.Remote)})
	c.NativeRunners, c.NativeLinuxPackages = s.Config.CI.NativeRunners, s.Config.CI.NativeLinuxPackages
	if s.Config.Release.GitHubReleases {
		c.Add(protocol.CIResult{Credentials: []protocol.Credential{{Env: "GITHUB_TOKEN", Description: "creates GitHub Releases and uploads their assets", Required: true}}})
	}
	return c, nil
}

// Rendered maps each configured provider's file to its content.
func Rendered(cfg workspace.Config, m Model) map[string]string {
	out := map[string]string{}
	for _, p := range cfg.CI.Providers {
		switch p {
		case "github":
			out[GitHubPath] = RenderGitHub(m)
			if m.Pub {
				out[GitHubPubPath] = RenderGitHubPub(m)
			}
		case "azure":
			m.AzurePool, m.AzureVariableGroups = cfg.CI.AzurePool, cfg.CI.AzureVariableGroups
			out[AzurePath] = RenderAzure(m)
		case "gitlab":
			out[GitLabPath] = RenderGitLab(m)
		}
	}
	if cfg.CI.Devcontainer {
		out[DevcontainerPath] = RenderDevcontainer(cfg.Workspace.Name, m)
	}
	return out
}

// Write renders and writes every pipeline, keeping slot contents. It returns
// the files that changed and any slot that could not be kept.
func Write(root string, cfg workspace.Config, m Model) (changed, lost []string, err error) {
	rendered := Rendered(cfg, m)
	for _, path := range Stale(root, rendered) {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			return changed, lost, err
		}
		changed = append(changed, path)
	}
	for path, content := range rendered {
		full := filepath.Join(root, filepath.FromSlash(path))
		old, readErr := os.ReadFile(full)
		if readErr != nil && !errors.Is(readErr, fs.ErrNotExist) {
			return changed, lost, readErr
		}
		content, missing := Preserve(string(old), content)
		for _, slot := range missing {
			lost = append(lost, path+": "+slot)
		}
		if string(old) == content {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return changed, lost, err
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return changed, lost, err
		}
		changed = append(changed, path)
	}
	return changed, lost, nil
}

// OwnedPaths are the files smith generates and may rewrite or delete. A file
// at one of them belongs to smith only while it starts with smith's header.
var OwnedPaths = []string{GitHubPath, GitHubPubPath, AzurePath, GitLabPath, DevcontainerPath}

// Stale lists the files smith generated that the workspace no longer asks for
// (a provider dropped from [ci] providers): they carry smith's header and are
// not rendered any more. A file without the header is the team's, and kept.
func Stale(root string, rendered map[string]string) []string {
	var out []string
	for _, path := range OwnedPaths {
		if _, wanted := rendered[path]; wanted {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err == nil && generated(string(raw)) {
			out = append(out, path)
		}
	}
	return out
}

func generated(text string) bool {
	return strings.HasPrefix(text, "# Generated by smith") || strings.HasPrefix(text, "// Generated by smith")
}

// Planned is what Write would change: each differing pipeline's new content,
// slot contents kept, and the slots that could not be kept.
func Planned(root string, cfg workspace.Config, m Model) (files map[string]string, lost []string) {
	files = map[string]string{}
	for path, content := range Rendered(cfg, m) {
		old, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		content, missing := Preserve(string(old), content)
		for _, slot := range missing {
			lost = append(lost, path+": "+slot)
		}
		if string(old) != content {
			files[path] = content
		}
	}
	return files, lost
}

// Drift lists pipelines whose committed content differs from what the
// workspace would generate now.
func Drift(root string, cfg workspace.Config, m Model) []string {
	var drift []string
	rendered := Rendered(cfg, m)
	for _, path := range Stale(root, rendered) {
		drift = append(drift, path+" is no longer configured")
	}
	for path, content := range rendered {
		old, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			drift = append(drift, path+" is missing")
			continue
		}
		if kept, _ := Preserve(string(old), content); kept != string(old) {
			drift = append(drift, path+" is out of date")
		}
	}
	return drift
}

// Regenerate collects contributions and rewrites the pipelines of an open
// workspace.
func Regenerate(s *session.Session) (changed, lost []string, err error) {
	c, err := Collect(s)
	if err != nil {
		return nil, nil, err
	}
	return Write(s.Root, s.Config, NewModel(c, s.Version, s.Config.Release.Branch))
}
