package pipeline

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// azureCLI installs the Azure CLI where it is missing (GitHub's and Azure's
// hosted images carry it) and the Bicep CLI it runs, on every provider.
var azureCLI = protocol.Toolchain{
	Name: "azure-cli", Version: "latest",
	Install: "command -v az >/dev/null || curl -fsSL https://aka.ms/InstallAzureCLIDeb | bash\n" +
		"az bicep install",
	Devcontainer: &protocol.DevcontainerFeature{ID: "ghcr.io/devcontainers/features/azure-cli:1", Options: map[string]string{"installBicep": "true"}},
}

// CI returns the Bicep contribution: the Azure CLI with Bicep.
func CI(params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	return protocol.CIResult{Toolchains: []protocol.Toolchain{azureCLI}}
}

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Probe(dir, argv) }

// Doctor checks that the Azure CLI and its Bicep CLI are installed.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, azErr := probe(workspaceRoot, []string{"az", "version"})
	findings := []protocol.Finding{{Check: "the Azure CLI is installed", OK: azErr == nil, Detail: "az is not on PATH", Remedy: "install the Azure CLI (https://aka.ms/installazurecli)"}}
	if azErr == nil {
		_, err := probe(workspaceRoot, []string{"az", "bicep", "version"})
		findings = append(findings, protocol.Finding{Check: "the Bicep CLI is installed", OK: err == nil, Detail: "az bicep version failed", Remedy: "run az bicep install"})
	}
	return protocol.DoctorResult{Findings: findings}
}
