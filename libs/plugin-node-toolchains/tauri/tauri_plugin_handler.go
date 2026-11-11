package tauri

import (
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/cargoworkspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/scaffold"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/versionsync"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
)

// Name is the plugin's name in a stack's toolchain list and in smith.toml
// ([conventions.tauri]).
const Name = tauriapp.Name

// Stack is the stack the plugin publishes.
const Stack = "node-vitest-tauri"

// Tauri is the tauri toolchain plugin.
type Tauri struct {
	Version string
	root    string
}

// New returns the plugin; version is reported in the handshake.
func New(version string) *Tauri { return &Tauri{Version: version} }

// Describe implements serve.Plugin.
func (p *Tauri) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	p.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.Version},
		Kind:      protocol.KindConvention,
		Languages: []string{"node"},
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
		Roles:     []protocol.Role{protocol.RoleToolchain},
		Provides:  []string{TaskName},
		Choices:   Choices(),
		// What a Tauri app needs on top of Node: Rust, and the system pieces of each system.
		Tools: toolcatalog.ForChoice(Name, "rust", "xcode-clt", "msvc-build-tools", "webview2", "tauri-linux-libs"),
		Stacks: []protocol.Stack{{
			Name:        Stack,
			Description: "TypeScript on Node, linted by eslint, unit-tested with Vitest, packaged as a desktop app with Tauri",
			Language:    "node",
			Scaffold:    "typescript",
			Toolchain:   []string{"typescript", "eslint", "vitest", Name},
		}},
	}, nil
}

// Contribute implements serve.Contributor.
func (p *Tauri) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	return Contribute(p.root, params), nil
}

// CI implements serve.CIContributor: the Rust toolchain and, on Linux, the system libraries.
func (p *Tauri) CI(params protocol.CIParams) (protocol.CIResult, error) {
	return CI(p.root, params), nil
}

// WriteVersion implements serve.Versioner: the version goes into src-tauri's
// Cargo.toml and tauri.conf.json, beside the package.json the node plugin writes.
func (p *Tauri) WriteVersion(params protocol.VersionParams) (protocol.VersionResult, error) {
	return versionsync.WriteVersion(p.root, params)
}

// Doctor implements serve.DoctorContributor.
func (p *Tauri) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	return Doctor(p.root, params), nil
}

// Generate implements serve.Generator: only a stack that names tauri gets the
// CLI and its script; as a convention the plugin acts on projects that have
// src-tauri and adds nothing. For a project with a web page it writes src-tauri
// (files already there stay) and keeps it out of the root Cargo workspace.
func (p *Tauri) Generate(params protocol.GenerateParams) (protocol.GenerateResult, error) {
	if params.Convention {
		return protocol.GenerateResult{}, nil
	}
	res, err := tooling.Apply(p.root, params, scaffold.Setup(p.root, params))
	if err != nil {
		return res, err
	}
	if tauriapp.Detect(p.root, params.Directory) {
		changed, err := scaffold.PointViteAtFrontend(p.root, params.Directory)
		if err != nil {
			return res, err
		}
		if changed != "" {
			res.Files = append(res.Files, changed)
		}
	}
	if excluded, err := cargoworkspace.ExcludeFromCargoWorkspace(p.root, params.Directory); err != nil {
		return res, err
	} else if excluded {
		res.Files = append(res.Files, "Cargo.toml")
	}
	return res, nil
}
