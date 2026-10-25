package recipe

import (
	"strings"
	"testing"
)

func loadedFrom(t *testing.T, text string) Loaded {
	t.Helper()
	doc, err := ParseRecipe([]byte(text), Source{Ref: "team.recipe.toml"})
	if err != nil {
		t.Fatal(err)
	}
	return Loaded{Chain: []Document{doc}}
}

const twoLanguages = `
name = "team"
version = "1.0.0"

[choices.language]
ask = "Language"
options = ["python", "node"]
default = "python"

[when.language.python.use]
language = "python"

[when.language.node.use]
language = "node"
lint = "eslint"
test = "jest"
`

func TestValidateAllChecksEveryAnswersBranch(t *testing.T) {
	n, err := ValidateAll(loadedFrom(t, twoLanguages), fakePlugins())
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("checked %d combinations, want 2 (python, node)", n)
	}
	// The default branch (python) is fine; node's is broken: two tests.
	broken := strings.Replace(twoLanguages, `test = "jest"`, "test = \"jest\"\nstructure = \"karma\"", 1)
	_, err = ValidateAll(loadedFrom(t, broken), fakePlugins())
	if err == nil {
		t.Fatal("a broken node branch was accepted, because the default branch is python")
	}
	if !strings.Contains(err.Error(), "with language=node") || !strings.Contains(err.Error(), "karma") {
		t.Errorf("err = %v; it should name the answers and the plugin", err)
	}
}

func TestValidateAllAsksChoicesThatHaveNoDefault(t *testing.T) {
	text := strings.Replace(twoLanguages, `default = "python"`, "", 1)
	if _, err := ValidateAll(loadedFrom(t, text), fakePlugins()); err != nil {
		t.Errorf("a choice without a default must be checked through its options: %v", err)
	}
}

func TestValidateAllCombinesIndependentChoices(t *testing.T) {
	text := twoLanguages + `
[choices.shipping]
options = ["none", "image"]
default = "none"

[when.shipping.image.use]
deploy = "docker"
structure = "docker"
`
	_, err := ValidateAll(loadedFrom(t, text), fakePlugins())
	if err == nil || !strings.Contains(err.Error(), "shipping=image") {
		t.Errorf("err = %v; docker is not a structure plugin, in the image branch", err)
	}
}
