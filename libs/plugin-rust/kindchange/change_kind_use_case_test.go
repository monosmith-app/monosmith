package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func edited(r protocol.ChangeKindResult) map[string]string {
	out := map[string]string{}
	for _, f := range r.Files {
		out[f.Path] = f.Content
	}
	return out
}

var workspace = map[string]string{
	"Cargo.toml": "[workspace]\nmembers = [\"libs/*\", \"apps/*\"]\n",
	"libs/core/Cargo.toml": "[package]\nname = \"core\"\nversion = \"0.1.0\"\npublish = false\n\n" +
		"[package.metadata.smith]\nkind = \"lib\"\n\n[dependencies]\nutil = { path = \"../util\" }\n",
	"libs/core/src/lib.rs":  "",
	"libs/util/Cargo.toml":  "[package]\nname = \"util\"\nversion = \"0.1.0\"\npublish = false\n\n[package.metadata.smith]\nkind = \"lib\"\n",
	"apps/cli/Cargo.toml":   "[package]\nname = \"cli\"\nversion = \"0.1.0\"\n\n[dependencies]\ncore = { path = \"../../libs/core\" }\n",
	"libs/other/Cargo.toml": "[package]\nname = \"other\"\nversion = \"0.1.0\"\n\n[dependencies]\ncore = { path = \"../core\", version = \"0.1\" }\n",
}

func projects() []protocol.Project {
	return []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "rust"},
		{Name: "util", Root: "libs/util", Language: "rust"},
		{Name: "cli", Root: "apps/cli", Language: "rust"},
		{Name: "other", Root: "libs/other", Language: "rust"},
	}
}

func TestLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: projects()[0], To: protocol.ProjectPackage, NewRoot: "packages/core", Projects: projects(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	core := got["libs/core/Cargo.toml"]
	if strings.Contains(core, "publish") || !strings.Contains(core, `kind = "package"`) {
		t.Fatalf("the publish guard goes and the kind changes:\n%s", core)
	}
	if !strings.Contains(core, `util = { path = "../../libs/util" }`) {
		t.Fatalf("its own path dependency still reaches util from the new directory:\n%s", core)
	}
	if !strings.Contains(got["apps/cli/Cargo.toml"], `path = "../../packages/core"`) {
		t.Fatalf("dependents follow:\n%s", got["apps/cli/Cargo.toml"])
	}
	if !strings.Contains(got["libs/other/Cargo.toml"], `path = "../../packages/core", version`) {
		t.Fatalf("a sibling follows:\n%s", got["libs/other/Cargo.toml"])
	}
	if !strings.Contains(got["Cargo.toml"], `"packages/core"`) {
		t.Fatalf("the workspace lists the new directory:\n%s", got["Cargo.toml"])
	}
	if _, touched := got["libs/util/Cargo.toml"]; touched {
		t.Fatal("a crate that does not depend on core is not edited")
	}
	joined := strings.Join(r.Warnings, "\n")
	if !strings.Contains(joined, "a description") || !strings.Contains(joined, "a license") {
		t.Fatalf("crates.io's requirements are named: %v", r.Warnings)
	}
}

func TestPackageBecomesLibAndGetsTheGuardBack(t *testing.T) {
	files := map[string]string{
		"Cargo.toml":               "[workspace]\nmembers = [\"packages/*\", \"libs/*\"]\n",
		"packages/sdk/Cargo.toml":  "[package]\nname = \"sdk\"\nversion = \"1.0.0\"\ndescription = \"x\"\nlicense = \"MIT\"\n\n[package.metadata.smith]\nkind = \"package\"\n",
		"packages/sdk/src/main.rs": "fn main() {}\n",
	}
	dir := writeFiles(t, files)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "sdk", Root: "packages/sdk", Language: "rust"},
		To:      protocol.ProjectLib, NewRoot: "libs/sdk",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)["packages/sdk/Cargo.toml"]
	if !strings.Contains(got, "publish = false") || !strings.Contains(got, `kind = "lib"`) {
		t.Fatalf("the guard is back and the kind is lib:\n%s", got)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "binary") {
		t.Fatalf("the binary is named: %v", r.Warnings)
	}
}

func TestLibToPackageAndBackLeavesTheManifestByteForByte(t *testing.T) {
	original := "[package]\nname = \"x\"\nversion = \"0.0.0\"\nedition = \"2021\"\npublish = false\n\n[package.metadata.smith]\nkind = \"lib\"\n\n[dependencies]\n"
	dir := writeFiles(t, map[string]string{"libs/x/Cargo.toml": original})
	project := protocol.Project{Name: "x", Root: "libs/x", Language: "rust"}
	for _, to := range []protocol.ProjectKind{protocol.ProjectPackage, protocol.ProjectLib} {
		r, err := Change(dir, protocol.ChangeKindParams{Project: project, To: to, NewRoot: "libs/x"})
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range r.Files {
			if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(e.Path)), []byte(e.Content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "libs/x/Cargo.toml")); string(got) != original {
		t.Fatalf("round trip changed the manifest:\n%s", got)
	}
}
