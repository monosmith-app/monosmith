package recipechoice

import (
	"strings"
	"testing"
)

func TestDescribeNamesTheRecipeTheAnswersAndWhatItUses(t *testing.T) {
	a, _ := asker("", false)
	res, err := a.Resolve(load(t), map[string]string{"language": "go"})
	if err != nil {
		t.Fatal(err)
	}
	got := Describe(res)
	for _, want := range []string{
		"Recipe team 1.2.0 (language=go, kind=library, cgo=no) will use:",
		"language",
		"go",
		"team.recipe.toml:",
		"setting go.cgo = maybe",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
}
