package plugin

import (
	"io"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/conformance"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// The plugin passes the suite `smith plugin check --generate` runs: it
// generates an app and a lib, discovers them back, writes their version.
func TestConforms(t *testing.T) {
	ws := t.TempDir()
	toPlugin, fromHost := io.Pipe()
	toHost, fromPlugin := io.Pipe()
	go func() { _ = serve.Serve(New("1.0.0"), toPlugin, fromPlugin) }()
	c, err := host.Connect(toHost, fromHost, host.Options{WorkspaceRoot: ws})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if report := conformance.Check(c, conformance.Options{Workspace: ws}); !report.OK() {
		t.Fatal(report.Failures())
	}
}
