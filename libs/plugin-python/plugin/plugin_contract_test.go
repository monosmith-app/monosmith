package plugin

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// The language contract (ADR 0053): the plugin declares channels and declared
// projects as decided, and a plugin that declares a capability can serve it.
func TestPluginDeclaresTheCapabilitiesItHonours(t *testing.T) {
	if _, ok := any(New("test")).(serve.ChannelPublisher); ok != true {
		t.Errorf("ChannelPublisher: want implemented")
	}
	if _, ok := any(New("test")).(serve.DeclaredDescriber); ok != true {
		t.Errorf("DeclaredDescriber: want implemented")
	}
}
