package tauri

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Choices are the choices the plugin publishes (InitializeResult.Choices): the
// desktop shell. It joins the toolchain of a web app that another choice writes
// (--with react,tauri), as the stack react-tauri does. For an app that exists,
// `smith tauri add` does the same.
func Choices() []protocol.Choice {
	return []protocol.Choice{{
		Axis: "shell", Name: "tauri", Title: "Tauri", Layer: "shell", Language: "node",
		Description: "wraps a web app in a small native desktop program (Rust and the system web view); adds src-tauri to the app. Needs Rust to build. For an app that exists, run smith tauri add",
		Aliases:     []string{"desktop", "windows", "mac", "linux", "native", "rust", "app"},
		Roles:       []protocol.Role{protocol.RoleToolchain},
		Conflicts:   []string{"electron", "azure-functions", "aws-lambda", "gcp-functions"},
	}}
}
