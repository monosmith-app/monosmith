package cipipeline

import (
	"slices"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// InstallScriptPath is where `smith ci generate` writes the smith installer
// inside the project; generated pipelines run that file with sh instead of
// fetching a script from a mutable branch on every run.
const InstallScriptPath = "tools/smith/install.sh"

// Model is a pipeline independent of any CI provider.
type Model struct {
	// SmithVersion is installed by the pipeline; "latest" when smith is a dev build.
	SmithVersion  string
	ReleaseBranch string
	// Branches are the branches the GitHub and Azure pipelines trigger on
	// (push and pull request), from [ci] branches. Empty means the release
	// branch only.
	Branches   []string
	Toolchains []protocol.Toolchain
	// Authenticate are the steps that sign in to private registries, run
	// after the toolchains and before setup.
	Authenticate []protocol.Toolchain
	Credentials  []protocol.Credential
	// Stores are the package stores the ci job caches, sorted by name.
	Stores []protocol.PackageStore
	// Pack is true when a project has a package task: pushes to the release
	// branch then run smith ci pack and keep dist/drop as the "drop" artifact.
	Pack bool
	// Native is true when a project is tagged native: a job per OS then runs
	// smith ci native.
	Native bool
	// NativePublish is true when a native package is published from those
	// jobs, which then need the release credentials too.
	NativePublish bool
	// NativeRunners and NativeLinuxPackages configure that job.
	NativeRunners       []string
	NativeLinuxPackages []string
	// Provenance is true when an app is published to npm: on GitHub the job
	// may then request an OIDC token, which npm provenance signs with.
	Provenance bool
	// Attest is true when a release attaches archives to a GitHub Release: on
	// GitHub the job then attests them (SLSA build provenance), which needs
	// id-token and attestations permissions.
	Attest bool
	// Pub is true when an app is published to pub.dev: on GitHub the release
	// then starts the pub workflow (actions: write), which publishes it.
	Pub bool
	// AzurePool and AzureVariableGroups are [ci] azure_pool and
	// azure_variable_groups, for the Azure Pipelines definition.
	AzurePool           string
	AzureVariableGroups []string
	// RubyGemsTrusted is true when gems are published with the workflow's
	// OIDC token: the release job then gets id-token permission and signs in
	// to RubyGems right before the release, and the key is not a secret.
	RubyGemsTrusted bool
	// VSCEEntra is true when extensions are published with Microsoft Entra
	// ID: the release job gets id-token permission and signs in to Azure
	// right before the release, and the Marketplace token is not a secret.
	VSCEEntra bool
	// Permissions are the GitHub token permissions plugins asked for
	// ("packages: write"), sorted.
	Permissions []string
}

// DefaultNativeRunners build native projects on each OS.
var DefaultNativeRunners = []string{"ubuntu-latest", "windows-latest", "macos-latest"}

// Contributions are every plugin's CI needs for this workspace.
type Contributions struct {
	Toolchains   []protocol.Toolchain
	Authenticate []protocol.Toolchain
	Setup        []protocol.Step
	Audit        []protocol.Step
	Stores       []protocol.PackageStore
	// PostRelease are the steps smith ci release runs after the release.
	PostRelease []protocol.Step
	Credentials []protocol.Credential
	// Pack is true when a project has a package task.
	Pack bool
	// Native is true when a project is tagged native.
	Native              bool
	NativePublish       bool
	NativeRunners       []string
	NativeLinuxPackages []string
	// Provenance is true when an app is published to npm.
	Provenance bool
	// Attest is true when a release attaches archives to a GitHub Release.
	Attest bool
	// Pub is true when an app is published to pub.dev.
	Pub bool
	// RubyGemsTrusted is true when [release] rubygems_trusted_publishing is set.
	RubyGemsTrusted bool
	// VSCEEntra is true when [release] vsce_entra is set.
	VSCEEntra bool
	// Permissions are the GitHub token permissions plugins asked for.
	Permissions []string
}

// UseRubyGemsTrustedPublishing makes gems publish with the workflow's OIDC
// token: the key is no longer a secret the release needs, because the
// pipeline signs in to RubyGems right before it.
func (c *Contributions) UseRubyGemsTrustedPublishing() {
	c.RubyGemsTrusted = true
	c.Credentials = slices.DeleteFunc(c.Credentials, func(cred protocol.Credential) bool { return cred.Env == "GEM_HOST_API_KEY" })
}

// UseVSCEEntra makes extensions publish with Microsoft Entra ID: the
// Marketplace token is no longer a secret the release needs, because the
// pipeline signs in to Azure right before it.
func (c *Contributions) UseVSCEEntra() {
	c.VSCEEntra = true
	c.Credentials = slices.DeleteFunc(c.Credentials, func(cred protocol.Credential) bool { return cred.Env == "VSCE_PAT" })
}

// Add merges one plugin's contribution, keeping the first of duplicates.
func (c *Contributions) Add(r protocol.CIResult) {
	for _, t := range r.Toolchains {
		if !hasToolchain(c.Toolchains, t.Name) {
			c.Toolchains = append(c.Toolchains, t)
		}
	}
	for _, a := range r.Authenticate {
		if !hasToolchain(c.Authenticate, a.Name) {
			c.Authenticate = append(c.Authenticate, a)
		}
	}
	for _, perm := range r.Permissions {
		if !slices.Contains(c.Permissions, perm) {
			c.Permissions = append(c.Permissions, perm)
		}
	}
	for _, s := range r.Stores {
		if !slices.ContainsFunc(c.Stores, func(have protocol.PackageStore) bool { return have.Name == s.Name }) {
			c.Stores = append(c.Stores, s)
		}
	}
	c.Setup = append(c.Setup, r.Setup...)
	c.Audit = append(c.Audit, r.Audit...)
	c.PostRelease = append(c.PostRelease, r.PostRelease...)
	for _, cred := range r.Credentials {
		dup := false
		for _, have := range c.Credentials {
			dup = dup || have.Env == cred.Env
		}
		if !dup {
			c.Credentials = append(c.Credentials, cred)
		}
		// A credential setup needs as well as the release keeps that need,
		// whichever plugin asked first.
		for i := range c.Credentials {
			if c.Credentials[i].Env == cred.Env && cred.Setup {
				c.Credentials[i].Setup = true
			}
			if c.Credentials[i].Env == cred.Env && cred.Tasks {
				c.Credentials[i].Tasks = true
			}
		}
	}
}

// TriggerBranches are the branches the pipelines trigger on: Branches, with the
// release branch first when they leave it out, or the release branch alone.
func (m Model) TriggerBranches() []string {
	out := []string{}
	if !slices.Contains(m.Branches, m.ReleaseBranch) {
		out = append(out, m.ReleaseBranch)
	}
	return append(out, m.Branches...)
}

// NewModel builds the model, sorted so the YAML is deterministic.
func NewModel(c Contributions, smithVersion, releaseBranch string) Model {
	if smithVersion == "" || smithVersion == "dev" {
		smithVersion = "latest"
	}
	m := Model{SmithVersion: smithVersion, ReleaseBranch: releaseBranch, Pack: c.Pack, Native: c.Native, NativePublish: c.NativePublish,
		NativeRunners: c.NativeRunners, NativeLinuxPackages: c.NativeLinuxPackages, Provenance: c.Provenance, Attest: c.Attest, Pub: c.Pub, RubyGemsTrusted: c.RubyGemsTrusted, VSCEEntra: c.VSCEEntra}
	if len(m.NativeRunners) == 0 {
		m.NativeRunners = DefaultNativeRunners
	}
	m.Permissions = append(m.Permissions, c.Permissions...)
	sort.Strings(m.Permissions)
	m.Authenticate = append(m.Authenticate, c.Authenticate...)
	sort.Slice(m.Authenticate, func(i, j int) bool { return m.Authenticate[i].Name < m.Authenticate[j].Name })
	m.Toolchains = append(m.Toolchains, c.Toolchains...)
	sort.Slice(m.Toolchains, func(i, j int) bool { return m.Toolchains[i].Name < m.Toolchains[j].Name })
	m.Stores = append(m.Stores, c.Stores...)
	sort.Slice(m.Stores, func(i, j int) bool { return m.Stores[i].Name < m.Stores[j].Name })
	m.Credentials = append(m.Credentials, c.Credentials...)
	sort.Slice(m.Credentials, func(i, j int) bool { return m.Credentials[i].Env < m.Credentials[j].Env })
	return m
}

func hasToolchain(list []protocol.Toolchain, name string) bool {
	for _, t := range list {
		if t.Name == name {
			return true
		}
	}
	return false
}

// TaskCredentials are the credentials every step that runs tasks needs: the
// remote cache's key and its backend's credentials.
func (m Model) TaskCredentials() []protocol.Credential {
	var out []protocol.Credential
	for _, c := range m.Credentials {
		if c.Tasks {
			out = append(out, c)
		}
	}
	return out
}

// nativeCredentials are the credentials the native job's step needs: those of
// the tasks it runs, and the release's when it publishes a native package.
func (m Model) nativeCredentials() []protocol.Credential {
	var out []protocol.Credential
	for _, c := range m.Credentials {
		if c.Tasks || (m.NativePublish && c.Env != "GITHUB_TOKEN") {
			out = append(out, c)
		}
	}
	return out
}

// SetupCredentials are the credentials setup needs too: a token installs
// read from a private registry.
func (m Model) SetupCredentials() []protocol.Credential {
	var out []protocol.Credential
	for _, c := range m.Credentials {
		if c.Setup {
			out = append(out, c)
		}
	}
	return out
}
