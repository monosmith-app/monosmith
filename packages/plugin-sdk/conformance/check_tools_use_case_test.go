package conformance

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
)

// toolser is a convention that declares the tools it is given.
type toolser struct{ tools []protocol.ToolNeed }

func (c toolser) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "toolser", Version: "1.0.0"}, Kind: protocol.KindConvention,
		Roles: []protocol.Role{protocol.RoleStructure}, Tools: c.tools}, nil
}

func (toolser) Contribute(protocol.ContributeParams) (protocol.ContributeResult, error) {
	return protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}, nil
}

func TestAPluginThatDeclaresWellFormedToolsConforms(t *testing.T) {
	r := Check(connect(t, toolser{tools: toolcatalog.For("python", "go")}, t.TempDir()), Options{})
	if !r.OK() {
		t.Fatalf("failures:\n%s", r.Failures())
	}
}

func TestAPluginThatDeclaresABadOrRepeatedToolIsReported(t *testing.T) {
	good, _ := toolcatalog.Find("go")
	bad := good
	bad.Name, bad.Install = "bad", []protocol.ToolInstall{{Manager: "chocolatey", Package: "x"}}
	for name, tools := range map[string][]protocol.ToolNeed{"bad manager": {bad}, "twice": {good, good}} {
		r := Check(connect(t, toolser{tools: tools}, t.TempDir()), Options{})
		if r.OK() || !strings.Contains(r.Failures(), "tool ") {
			t.Errorf("%s: want a failure about the tool, got:\n%s", name, r.Failures())
		}
	}
}
