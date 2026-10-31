package conformance

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// chooser is a convention that publishes the choices it is given.
type chooser struct{ choices []protocol.Choice }

func (c chooser) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "chooser", Version: "1.0.0"}, Kind: protocol.KindConvention,
		Roles: []protocol.Role{protocol.RoleStructure}, Choices: c.choices}, nil
}

func (chooser) Contribute(protocol.ContributeParams) (protocol.ContributeResult, error) {
	return protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}, nil
}

func TestAPluginThatPublishesWellFormedChoicesConforms(t *testing.T) {
	r := Check(connect(t, chooser{choices: []protocol.Choice{{Axis: "style", Name: "bootstrap", Language: "node"}}}, t.TempDir()), Options{})
	if !r.OK() {
		t.Fatalf("failures:\n%s", r.Failures())
	}
}

func TestAPluginThatPublishesABadOrRepeatedChoiceIsReported(t *testing.T) {
	good := protocol.Choice{Axis: "style", Name: "bootstrap", Language: "node"}
	for name, choices := range map[string][]protocol.Choice{
		"bad":      {{Axis: "Style", Name: "bootstrap", Language: "node"}},
		"repeated": {good, good},
	} {
		failures := Check(connect(t, chooser{choices: choices}, t.TempDir()), Options{}).Failures()
		if !strings.Contains(failures, "choice bootstrap is well formed") {
			t.Errorf("%s not reported:\n%s", name, failures)
		}
	}
}
