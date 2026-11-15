package recipe

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCatalogDescribesBuiltInPluginsWithoutStartingThem(t *testing.T) {
	c := Catalog{Version: "9.9.9"}
	node, err := c.Lookup("node")
	if err != nil || node.Plugin.Version != "9.9.9" || !slices.Contains(node.EffectiveRoles(), protocol.RoleLanguage) || !node.Capabilities.Generate {
		t.Fatalf("node = %+v, %v", node, err)
	}
	eslint, err := c.Lookup("eslint")
	if err != nil || !slices.Contains(eslint.Provides, "lint") || len(eslint.Requires) == 0 {
		t.Fatalf("eslint = %+v, %v", eslint, err)
	}
}

func TestCatalogSaysWhatIsKnownWhenAPluginIsNot(t *testing.T) {
	_, err := Catalog{Root: t.TempDir(), Version: "1.0.0"}.Lookup("pytest")
	if !errors.Is(err, ErrUnknownPlugin) || !strings.Contains(err.Error(), "built in:") || !strings.Contains(err.Error(), "vertical-slices") {
		t.Errorf("err = %v", err)
	}
}

func TestRealBuiltInPluginsAcceptTheTeamRecipe(t *testing.T) {
	team := mustParse(t, "team.toml", teamRecipe)
	for _, answers := range []map[string]string{
		{"language": "python"},
		{"language": "csharp"},
		{"language": "csharp", "kind": "service"},
	} {
		r, err := Resolve(chain(team), answers)
		if err != nil {
			t.Fatal(err)
		}
		if err := Validate(r, Catalog{Version: "1.0.0"}.Lookup); err != nil {
			t.Errorf("%v: %v", answers, err)
		}
	}
}
