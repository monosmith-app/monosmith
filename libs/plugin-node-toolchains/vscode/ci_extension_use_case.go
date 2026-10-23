package vscode

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ciExtensions is what a pipeline needs to release extension apps: the
// Marketplace token (which [release] vsce_entra replaces with an Azure sign-in)
// and, optionally, the Open VSX token.
func ciExtensions(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	var result protocol.CIResult
	for _, p := range params.Projects {
		if !isExtensionApp(workspaceRoot, p) {
			continue
		}
		result.Credentials = []protocol.Credential{
			{Env: "VSCE_PAT", Description: "Visual Studio Marketplace token that can publish the workspace's extensions", Required: true},
			{Env: "OVSX_PAT", Description: "Open VSX token: extensions are also published to open-vsx.org when it is set", Required: false},
		}
		break
	}
	return result
}
