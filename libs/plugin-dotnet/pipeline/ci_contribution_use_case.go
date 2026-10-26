package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/registries"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// SDKVersion is installed when global.json does not pin one; bump it
// deliberately.
const SDKVersion = "10.0"

// Version is the SDK global.json pins (its major.minor channel), else
// SDKVersion.
func Version(workspaceRoot string) string {
	var g struct {
		SDK struct{ Version string } `json:"sdk"`
	}
	if raw, err := os.ReadFile(filepath.Join(workspaceRoot, "global.json")); err == nil && json.Unmarshal(raw, &g) == nil && g.SDK.Version != "" {
		parts := strings.SplitN(g.SDK.Version, ".", 3)
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1]
		}
	}
	return SDKVersion
}

func runtime(version string) protocol.Toolchain {
	return protocol.Toolchain{
		Name: "dotnet", Version: version,
		GitHub: &protocol.ActionStep{Uses: "actions/setup-dotnet@v4", Inputs: []protocol.Input{{Name: "dotnet-version", Value: version + ".x"}}},
		Azure:  &protocol.ActionStep{Uses: "UseDotNet@2", Inputs: []protocol.Input{{Name: "packageType", Value: "sdk"}, {Name: "version", Value: version + ".x"}}},
		// Elsewhere (GitLab, on a Debian image): Microsoft's install script.
		Install: "curl -fsSL https://dot.net/v1/dotnet-install.sh | bash -s -- --channel " + version + " --install-dir /usr/local/dotnet\n" +
			"export PATH=/usr/local/dotnet:$PATH DOTNET_ROOT=/usr/local/dotnet",
		Devcontainer: &protocol.DevcontainerFeature{ID: "ghcr.io/devcontainers/features/dotnet:2", Options: map[string]string{"version": version}},
	}
}

// CI returns the .NET contribution: the SDK, a restore of each project (its
// tests project too), the sign-in a private feed needs, and the NuGet key
// when a package is pushed to nuget.org.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	result := protocol.CIResult{
		Toolchains: []protocol.Toolchain{runtime(Version(workspaceRoot))},
		Stores:     []protocol.PackageStore{{Name: "nuget", Path: "~/.nuget/packages", Env: "NUGET_PACKAGES", Files: []string{"**/*.csproj", "**/*.fsproj"}}},
	}
	signIn, publicPublish := registries.CI(params.Projects)
	result.Authenticate = signIn.Authenticate
	result.Credentials = append(result.Credentials, signIn.Credentials...)
	result.Permissions = signIn.Permissions
	for _, p := range params.Projects {
		result.Setup = append(result.Setup, protocol.Step{Name: "dotnet restore " + p.Name, Command: []string{"dotnet", "restore"}, Dir: p.Root, Blocking: true})
		if t, ok := p.Tasks["test"]; ok && len(t.Command) > 2 {
			result.Setup = append(result.Setup, protocol.Step{Name: "dotnet restore " + p.Name + " tests", Command: []string{"dotnet", "restore", t.Command[2]}, Dir: p.Root, Blocking: true})
		}
		if p.Publishes() && publicPublish && !hasCredential(result.Credentials, "NUGET_API_KEY") {
			result.Credentials = append(result.Credentials, protocol.Credential{Env: "NUGET_API_KEY", Description: "NuGet API key that can push the workspace's packages", Required: true})
		}
	}
	return result
}

func hasCredential(list []protocol.Credential, env string) bool {
	for _, c := range list {
		if c.Env == env {
			return true
		}
	}
	return false
}

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// Doctor checks that the .NET SDK is installed.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	out, err := probe(workspaceRoot, []string{"dotnet", "--version"})
	findings := []protocol.Finding{{
		Check: "the .NET SDK is installed", OK: err == nil,
		Detail: "dotnet is not on PATH", Remedy: "install the .NET SDK " + Version(workspaceRoot) + " (https://dot.net)",
	}}
	if f, ok := toolrun.CIPinFinding(".NET SDK", strings.TrimSpace(out), Version(workspaceRoot)); ok && err == nil {
		f.Remedy = "install the .NET SDK " + Version(workspaceRoot) + " (https://dot.net), or pin the SDK in global.json and run `smith ci generate` to move CI and the dev container"
		findings = append(findings, f)
	}
	if f, ok := registries.CheckNuGetConfig(workspaceRoot, params.Projects); ok {
		findings = append(findings, f)
	}
	return protocol.DoctorResult{Findings: findings}
}
