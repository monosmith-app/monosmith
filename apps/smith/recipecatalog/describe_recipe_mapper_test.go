package recipecatalog

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

const teamRecipe = `name = "team"
version = "1.2.0"

[choices.language]
ask = "Language"
options = ["python", "csharp"]

[choices.kind]
ask = "What are you creating?"
options = ["library", "service"]
default = "library"

[use]
structure = "vertical-slices"

[when.language.python.use]
language = "python"

[when.language.csharp.use]
language = "dotnet"

[when.language.csharp.settings.dotnet]
target_framework = "net10.0"

[when.language.csharp.when.kind.service.use]
deploy = "docker"
`

func TestDescribeListsChoicesRolesAndConfiguredPluginsWithoutAnswering(t *testing.T) {
	file := filepath.Join(t.TempDir(), "team.recipe.toml")
	if err := os.WriteFile(file, []byte(teamRecipe), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := (&recipe.Loader{}).Load(file)
	if err != nil {
		t.Fatal(err)
	}
	got := Describe(loaded)
	if got.Name != "team" || got.Version != "1.2.0" || len(got.Chain) != 1 {
		t.Fatalf("described %+v", got)
	}
	if len(got.Choices) != 2 || got.Choices[0].Name != "kind" || got.Choices[0].Default != "library" || got.Choices[0].Origin.Line == 0 {
		t.Fatalf("choices %+v", got.Choices)
	}
	roles := map[string]DescribedUse{}
	for _, u := range got.Uses {
		roles[u.Key] = u
	}
	if !reflect.DeepEqual(roles["structure"].Plugins, []string{"vertical-slices"}) {
		t.Errorf("structure %+v", roles["structure"])
	}
	wantLanguage := map[string][]string{"language=python": {"python"}, "language=csharp": {"dotnet"}}
	if !reflect.DeepEqual(roles["language"].Alternatives, wantLanguage) {
		t.Errorf("language %+v", roles["language"])
	}
	if !reflect.DeepEqual(roles["deploy"].Alternatives, map[string][]string{"language=csharp,kind=service": {"docker"}}) {
		t.Errorf("deploy %+v", roles["deploy"])
	}
	if !reflect.DeepEqual(got.Configures, []string{"dotnet"}) {
		t.Errorf("configures %v", got.Configures)
	}
}
