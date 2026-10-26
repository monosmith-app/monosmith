package pipeline

import (
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/registries"
	"github.com/monosmith-app/monosmith/libs/plugin-node/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CI returns the Node contribution to every pipeline, or nothing when the
// workspace has no Node project. Publishing needs NODE_AUTH_TOKEN for a
// package published to npmjs.org and VSCE_PAT for a VS Code extension; a
// private registry signs in its own way (registries.CI).
func CI(workspaceRoot string, params protocol.CIParams, nodeVersion string) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	result := protocol.CIResult{
		Toolchains: []protocol.Toolchain{runtime(nodeVersion)},
		Stores:     []protocol.PackageStore{{Name: "npm", Path: "~/.npm", Env: "NPM_CONFIG_CACHE", Files: []string{"package-lock.json"}}},
		Setup:      []protocol.Step{{Name: "Install Node dependencies", Command: []string{"npm", "ci", "--no-audit", "--no-fund"}, Blocking: true}},
		// smith npm-audit (AuditNpm) fails only on a production advisory npm
		// can fix; the rest are notes.
		Audit: []protocol.Step{{Name: "npm audit (runtime dependencies)", Command: []string{"{smith}", "npm-audit"}, Blocking: true, CacheFiles: []string{"package-lock.json"}}},
	}
	if params.Audit.Enabled("node") {
		// Opt-in ([ci] audit_policies): lockfile integrity, no git/http
		// dependencies, install scripts only from the allow-list.
		command := []string{"{smith}", "npm-policy"}
		if allowed := AllowedInstallScripts(workspaceRoot, params.Audit.AllowInstallScripts); len(allowed) > 0 {
			command = append(command, "--allow-scripts="+strings.Join(allowed, ","))
		}
		result.Audit = append(result.Audit, protocol.Step{Name: "npm policy (lockfile)", Command: command, Blocking: params.Audit.Blocks("node"), CacheFiles: []string{"package-lock.json"}})
	}
	signIn, publicPublish := registries.CI(params.Projects)
	result.Authenticate = signIn.Authenticate
	result.Credentials = append(result.Credentials, signIn.Credentials...)
	result.Permissions = append(result.Permissions, signIn.Permissions...)
	npm, marketplace := false, false
	for _, p := range params.Projects {
		if !p.Publishes() {
			continue
		}
		if _, ok := releasing.Extension(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))); ok {
			marketplace = true
		} else {
			npm = true
		}
	}
	if npm && publicPublish {
		result.Credentials = append(result.Credentials, protocol.Credential{Env: "NODE_AUTH_TOKEN", Description: "npm token that can publish the workspace's packages", Required: true})
	}
	if marketplace {
		result.Credentials = append(result.Credentials, protocol.Credential{Env: "VSCE_PAT", Description: "Visual Studio Marketplace token that can publish the workspace's extensions", Required: true})
	}
	return result
}
