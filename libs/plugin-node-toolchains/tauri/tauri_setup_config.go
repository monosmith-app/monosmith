package tauri

import (
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// cliVersion is the @tauri-apps/cli a new project starts with.
const cliVersion = "^2.0.0"

func setup(_ string, _ protocol.GenerateParams) tooling.Setup {
	return tooling.Setup{
		DevDependencies: map[string]string{"@tauri-apps/cli": cliVersion},
		Scripts:         map[string]string{"tauri": "tauri"},
	}
}
