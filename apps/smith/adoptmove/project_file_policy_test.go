package adoptmove

import (
	"reflect"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		eco, name string
		want      Role
	}{
		{EcosystemNode, "src", RoleProject},
		{EcosystemNode, "package.json", RoleProject},
		{EcosystemNode, "tsconfig.json", RoleProject},
		{EcosystemNode, "Dockerfile", RoleProject},
		{EcosystemNode, ".env", RoleProject},
		{EcosystemNode, "package-lock.json", RoleRepository},
		{EcosystemNode, "yarn.lock", RoleRepository},
		{EcosystemGo, "package-lock.json", RoleProject},
		{EcosystemGo, "go.sum", RoleProject},
		{EcosystemGo, "cmd", RoleProject},
		{EcosystemPython, "poetry.lock", RoleProject},
		{EcosystemPython, "requirements.txt", RoleProject},
		{EcosystemRust, "Cargo.lock", RoleProject},
		{EcosystemRust, ".cargo", RoleProject},
		{EcosystemNode, ".git", RoleRepository},
		{EcosystemNode, ".github", RoleRepository},
		{EcosystemGo, ".gitignore", RoleRepository},
		{EcosystemGo, "README.md", RoleRepository},
		{EcosystemGo, "LICENSE-MIT", RoleRepository},
		{EcosystemGo, "CHANGELOG.md", RoleRepository},
		{EcosystemGo, "docs", RoleRepository},
		{EcosystemGo, ".vscode", RoleRepository},
		{EcosystemGo, "CLAUDE.md", RoleRepository},
		{EcosystemGo, "smith.toml", RoleRepository},
		{EcosystemGo, "docker-compose.yml", RoleRepository},
		{EcosystemGo, "azure-pipelines.yml", RoleRepository},
	}
	for _, c := range cases {
		if got := Classify(c.eco, c.name); got != c.want {
			t.Errorf("Classify(%s, %q) = %v, want %v", c.eco, c.name, got, c.want)
		}
	}
}

func TestTopLevelKeepsFirstSegmentsOnce(t *testing.T) {
	got := TopLevel([]string{"src/a.js", "src/b/c.js", "package.json", ".github/ci.yml"})
	want := []string{".github", "package.json", "src"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
