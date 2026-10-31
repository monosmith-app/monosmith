package tauriadd

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRenderStackLoadsAsSmithToml(t *testing.T) {
	st := protocol.Stack{Name: "node-jest-tauri", Description: "Node, with Tauri", Language: "node", Toolchain: []string{"eslint", "jest", "tauri"}, Scaffold: "typescript"}
	text := "[workspace]\nname = \"ws\"\n\n" + renderStack(st, "studio")
	if _, err := workspace.Parse([]byte(text)); err != nil {
		t.Fatalf("%v\n%s", err, text)
	}
	if !strings.Contains(text, `toolchain = ["eslint", "jest", "tauri"]`) || !strings.Contains(text, "[stacks.node-jest-tauri]") {
		t.Errorf("text:\n%s", text)
	}
}
