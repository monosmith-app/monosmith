package recipestack

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestNameOfIsTheRecipeAndItsAnswers(t *testing.T) {
	if got := NameOf(python()); got != "team-python-library" {
		t.Errorf("got %q", got)
	}
	res := recipe.Resolution{Name: "Acme Team!", Asked: []recipe.Question{{Name: "x", Answer: "C#"}, {Name: "y", Answer: ".NET 10"}}}
	if got := NameOf(res); got != "acme-team-c-net-10" {
		t.Errorf("got %q", got)
	}
}

func TestChooseReusesTheSameStackAndSuffixesADifferentOne(t *testing.T) {
	st := protocol.Stack{Language: "python", Toolchain: []string{"ruff"}}
	none := func(string) (protocol.Stack, bool) { return protocol.Stack{}, false }
	if name, reuse, err := Choose("team-python", st, none); err != nil || reuse || name != "team-python" {
		t.Fatalf("new: %q %v %v", name, reuse, err)
	}
	same := func(string) (protocol.Stack, bool) {
		return protocol.Stack{Name: "team-python", Description: "other text", Language: "python", Toolchain: []string{"ruff"}, Structure: []string{}}, true
	}
	if name, reuse, err := Choose("team-python", st, same); err != nil || !reuse || name != "team-python" {
		t.Fatalf("same: %q %v %v", name, reuse, err)
	}
	other := func(n string) (protocol.Stack, bool) {
		if n == "team-python" {
			return protocol.Stack{Language: "python", Toolchain: []string{"flake8"}}, true
		}
		return protocol.Stack{}, false
	}
	name, reuse, err := Choose("team-python", st, other)
	if err != nil || reuse || !strings.HasPrefix(name, "team-python-") || len(name) != len("team-python-")+6 {
		t.Fatalf("different: %q %v %v", name, reuse, err)
	}
	if again, _, _ := Choose("team-python", st, other); again != name {
		t.Errorf("the suffix is not stable: %q and %q", name, again)
	}
}
