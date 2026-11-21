package configedit_test

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/features/pluginhost"
)

// plugins.enabled takes the names of the built-in language plugins only (smith
// refuses any other when it starts), so the schema lists them as its options.
// This fails when a built-in plugin is added or removed and the schema is not.
func TestPluginsEnabledOffersExactlyTheBuiltInLanguagePlugins(t *testing.T) {
	f, err := configedit.Find("plugins.enabled")
	if err != nil {
		t.Fatal(err)
	}
	var offered []string
	for _, o := range f.Options {
		offered = append(offered, o.Value)
		if !o.Default {
			t.Errorf("%s is loaded when plugins.enabled is not set, so it is a default", o.Value)
		}
	}
	if want := pluginhost.BuiltinNames(); !slices.Equal(offered, want) {
		t.Fatalf("plugins.enabled offers %v, the built-in plugins are %v: update the items enum and enumDescriptions in smith.schema.json", offered, want)
	}
}
