package sbom

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func put(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(abs), 0o755)
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListReadsTheUVLockAndSkipsTheProject(t *testing.T) {
	root := t.TempDir()
	put(t, root, "uv.lock", "version = 1\n\n[[package]]\nname = \"app\"\nversion = \"0.1.0\"\nsource = { editable = \".\" }\n\n[[package]]\nname = \"Typing_Extensions\"\nversion = \"4.12.2\"\nsource = { registry = \"https://pypi.org/simple\" }\n")
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "apps/app"}})
	if err != nil || res.Source != "uv.lock" || len(res.Components) != 1 || res.Components[0].PURL != "pkg:pypi/typing-extensions@4.12.2" {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestListFallsBackToRequirementsThenPyproject(t *testing.T) {
	root := t.TempDir()
	put(t, root, "requirements.txt", "# pins\nrequests==2.31.0\nflask>=3,<4 ; python_version > '3.8'\n-r other.txt\nurllib3[socks] (==2.0.0)\n")
	res, err := List(root, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || res.Source != "requirements.txt" || len(res.Components) != 3 {
		t.Fatalf("%+v %v", res, err)
	}
	want := map[string]bool{"pkg:pypi/requests@2.31.0": true, "pkg:pypi/flask@%3E=3%2C%3C4": true, "pkg:pypi/urllib3@2.0.0": true}
	for _, c := range res.Components {
		if !want[c.PURL] {
			t.Errorf("unexpected %s", c.PURL)
		}
	}

	other := t.TempDir()
	put(t, other, "pyproject.toml", "[project]\nname = \"x\"\ndependencies = [\"httpx>=0.27\"]\n")
	res, err = List(other, protocol.ComponentsParams{Project: protocol.Project{Root: "."}})
	if err != nil || res.Source != "pyproject.toml" || len(res.Components) != 1 || res.Components[0].Name != "httpx" {
		t.Fatalf("%+v %v", res, err)
	}
}
