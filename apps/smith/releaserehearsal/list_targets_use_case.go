package releaserehearsal

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pypiwheels"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ListTargets lists the registries the workspace's packages publish to: one
// target per package project whose plugin publishes to npm, PyPI, NuGet or
// crates.io. A project with `registry = "none"` has none. The URL is the
// registry smith.toml names, else what the publishing tool itself reads
// (NPM_CONFIG_REGISTRY and .npmrc, TWINE_REPOSITORY_URL, NUGET_SOURCE), else
// the public registry.
func ListTargets(s *session.Session, env func(string) string) []Target {
	var out []Target
	for _, p := range s.Graph.Projects {
		if p.Kind != protocol.ProjectPackage || (p.Registries != nil && p.Registries.NoPublish) {
			continue
		}
		names := p.Exports
		if len(names) == 0 {
			names = []string{p.Name}
		}
		var reg *protocol.Registry
		if p.Registries != nil {
			reg = p.Registries.Publish
		}
		t := Target{Project: p.Name, Names: names}
		if reg != nil {
			t.Auth, t.TokenEnv = reg.Auth, reg.TokenEnv
		}
		switch p.Plugin {
		case "node":
			t.Ecosystem = NPM
			t.URL = npmURL(s.Root, p.Root, names[0], reg, env)
			rc := readNpmrc(filepath.Join(userHome(env), ".npmrc"), filepath.Join(s.Root, ".npmrc"), filepath.Join(s.Root, filepath.FromSlash(p.Root), ".npmrc"))
			t.NpmrcToken, t.NpmrcVar = NpmrcToken(rc, t.URL, env)
		case "python":
			t.Ecosystem = PyPI
			t.URL = first(regField(reg, func(r *protocol.Registry) string { return r.PyPIUpload }), env("TWINE_REPOSITORY_URL"), pypiwheels.DefaultUploadURL)
		case "dotnet":
			t.Ecosystem = NuGet
			t.URL = first(regField(reg, func(r *protocol.Registry) string { return r.NuGet }), env("NUGET_SOURCE"), "https://api.nuget.org/v3/index.json")
		case "rust":
			t.Ecosystem = Crates
			t.URL = "https://crates.io"
		default:
			continue
		}
		out = append(out, t)
	}
	return out
}

func npmURL(root, project, name string, reg *protocol.Registry, env func(string) string) string {
	if reg != nil && reg.NPM != "" {
		return reg.NPM
	}
	if v := env("NPM_CONFIG_REGISTRY"); v != "" {
		return v
	}
	rc := readNpmrc(filepath.Join(userHome(env), ".npmrc"), filepath.Join(root, ".npmrc"), filepath.Join(root, filepath.FromSlash(project), ".npmrc"))
	return first(NpmrcRegistry(rc, name), "https://registry.npmjs.org")
}

func userHome(env func(string) string) string {
	if h := env("HOME"); h != "" {
		return h
	}
	if h := env("USERPROFILE"); h != "" {
		return h
	}
	h, _ := os.UserHomeDir()
	return h
}

func regField(r *protocol.Registry, f func(*protocol.Registry) string) string {
	if r == nil {
		return ""
	}
	return f(r)
}

func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
