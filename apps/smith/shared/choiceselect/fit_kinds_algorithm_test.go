package choiceselect

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFitKindsKeepsAFrameworkOffAKindItCannotWrite(t *testing.T) {
	kinds := map[string][]protocol.ProjectKind{
		"vite":   {protocol.ProjectApp, protocol.ProjectLib},
		"kit":    {protocol.ProjectApp},
		"pip":    {protocol.ProjectApp, protocol.ProjectLib, protocol.ProjectPackage},
		"eslint": nil,
	}
	got := FitKinds(webOffers(), func(plugin string) []protocol.ProjectKind { return kinds[plugin] })
	want := map[string][]protocol.ProjectKind{
		"node/react":     {protocol.ProjectApp, protocol.ProjectLib},
		"node/kit":       {protocol.ProjectApp},
		"node/bootstrap": {protocol.ProjectApp, protocol.ProjectLib},
		"node/chakra":    {protocol.ProjectApp, protocol.ProjectLib},
	}
	for key, kinds := range want {
		if !slices.Equal(got[key], kinds) {
			t.Errorf("%s: kinds %v, want %v", key, got[key], kinds)
		}
	}
	for _, key := range []string{"node/eslint", "node/jest", "node/vitest"} {
		if _, listed := got[key]; listed {
			t.Errorf("%s applies on any project, so it has no kinds: %v", key, got[key])
		}
	}
}
