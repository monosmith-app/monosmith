package linking

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAddRequirementListsTheDistributionOnce(t *testing.T) {
	src := "[project]\nname = \"api\"\nversion = \"0.1.0\"\ndependencies = [\n    \"requests>=2\",\n]\n"
	got, changed := AddRequirement(src, "shop-core-py", "")
	if !changed || !strings.Contains(got, "\"requests>=2\"") || !strings.Contains(got, "\"shop-core-py\"") {
		t.Fatalf("got %q %v", got, changed)
	}
	again, changed := AddRequirement(got, "Shop_Core.py", "")
	if changed || again != got {
		t.Fatalf("a name PyPI treats as the same is already listed: %v", changed)
	}
}

func TestAddRequirementCreatesTheKeyAndKeepsASpec(t *testing.T) {
	got, changed := AddRequirement("[project]\nname = \"api\"\n", "shop-core-py", ">=1.0")
	if !changed || !strings.Contains(got, "dependencies = [\"shop-core-py>=1.0\"]") {
		t.Fatalf("got %q", got)
	}
}

func TestRemoveRequirement(t *testing.T) {
	src := "[project]\nname = \"api\"\ndependencies = [\"shop-core-py\", \"requests\"]\n"
	got, changed := RemoveRequirement(src, "shop-core-py")
	if !changed || strings.Contains(got, "shop-core-py") || !strings.Contains(got, "requests") {
		t.Fatalf("got %q", got)
	}
	if _, changed := RemoveRequirement(got, "shop-core-py"); changed {
		t.Fatal("removing what is not listed changes nothing")
	}
}

func TestPlanNamesTheDistributionNotTheImportName(t *testing.T) {
	root := t.TempDir()
	for name, content := range map[string]string{
		"apps/api/pyproject.toml":     "[project]\nname = \"api\"\n",
		"libs/core-py/pyproject.toml": "[project]\nname = \"shop-core-py\"\nversion = \"1.0.0\"\n",
	} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	res, err := Link(root, protocol.LinkParams{
		Plan:    true,
		Project: protocol.Project{Name: "api", Root: "apps/api"},
		Library: protocol.Project{Name: "core-py", Root: "libs/core-py", Exports: []string{"shop_core_py"}},
	})
	if err != nil || res.Manifest != "apps/api/pyproject.toml" || !strings.Contains(res.Line, "shop-core-py") {
		t.Fatalf("got %+v %v", res, err)
	}
	if raw, _ := os.ReadFile(filepath.Join(root, "apps/api/pyproject.toml")); strings.Contains(string(raw), "shop-core") {
		t.Fatal("a plan writes nothing")
	}
}

func TestLinkNamesTheMissingManifest(t *testing.T) {
	_, err := Link(t.TempDir(), protocol.LinkParams{Project: protocol.Project{Name: "gone", Root: "apps/gone"}, Library: protocol.Project{Root: "libs/x"}})
	if err == nil || !strings.Contains(err.Error(), "apps/gone/pyproject.toml") {
		t.Fatalf("got %v", err)
	}
}
