package conformance

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// publisher is a convention that publishes the facts it is given.
type publisher struct{ facts map[string]any }

func (publisher) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "pub", Version: "1.0.0"}, Kind: protocol.KindConvention,
		Roles: []protocol.Role{protocol.RoleStructure}}, nil
}

func (p publisher) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, pr := range params.Projects {
		res.Projects[pr.Name] = protocol.Contribution{Facts: p.facts}
	}
	return res, nil
}

func TestAPluginThatPublishesValidFactsConforms(t *testing.T) {
	r := Check(connect(t, publisher{facts: map[string]any{protocol.FactSourceRoots: []string{"src"}}}, t.TempDir()), Options{})
	if !r.OK() {
		t.Fatalf("failures:\n%s", r.Failures())
	}
}

func TestAPluginThatPublishesAnInvalidFactIsReported(t *testing.T) {
	failures := Check(connect(t, publisher{facts: map[string]any{protocol.FactSourceRoots: "src", "nokey": 1}}, t.TempDir()), Options{}).Failures()
	if !strings.Contains(failures, "contribute publishes valid facts") {
		t.Errorf("not reported:\n%s", failures)
	}
}
