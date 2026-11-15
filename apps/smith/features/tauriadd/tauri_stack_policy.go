package tauriadd

import (
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const tauriPlugin = "tauri"

// DeriveStack is the stack a project has once it wraps a Tauri app: its current
// stack with tauri added last to the toolchain, so its eslint, test and web
// build tooling stay as they are. lookup finds a stack by name. A stack that
// already has that make-up is reused (react becomes react-tauri); otherwise a
// new one is named after the current stack, or after the project when that
// name is taken. declare says the stack must be written to smith.toml.
func DeriveStack(current protocol.Stack, project string, lookup func(string) (protocol.Stack, bool)) (st protocol.Stack, declare bool) {
	if slices.Contains(current.Toolchain, tauriPlugin) {
		return current, false
	}
	base, baseName, description := current, current.Name, current.Description
	if current.Name == "" || current.Name == protocol.NoStack {
		base, baseName, description = protocol.Stack{Language: "node"}, "node", "Node"
	}
	want := base
	want.Toolchain = append(slices.Clone(base.Toolchain), tauriPlugin)
	want.Description = description + ", packaged as a desktop app with Tauri"
	for _, name := range []string{baseName + "-tauri", project + "-tauri"} {
		want.Name = name
		existing, found := lookup(name)
		if !found {
			return want, true
		}
		if sameMakeUp(existing, want) {
			return existing, false
		}
	}
	return want, true
}

func sameMakeUp(a, b protocol.Stack) bool {
	return a.Language == b.Language && a.Scaffold == b.Scaffold &&
		slices.Equal(a.Toolchain, b.Toolchain) && slices.Equal(a.Structure, b.Structure) && slices.Equal(a.Deploy, b.Deploy)
}
