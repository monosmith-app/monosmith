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

const coreManifest = `{
    "name": "acme/core",
    "description": "core",
    "type": "library",
    "license": "MIT",
    "version": "0.0.0",
    "autoload": {
        "psr-4": {
            "Acme\\Core\\": "src/"
        }
    },
    "extra": {
        "smith": {
            "kind": "lib"
        }
    }
}
`

var workspace = map[string]string{
	"composer.json":           "{\n    \"name\": \"acme/workspace\",\n    \"repositories\": {\n        \"libs-core\": {\n            \"type\": \"path\",\n            \"url\": \"libs/core\"\n        },\n        \"apps-cli\": {\n            \"type\": \"path\",\n            \"url\": \"apps/cli\"\n        }\n    },\n    \"require\": {\n        \"acme/core\": \"@dev\"\n    }\n}\n",
	"composer.lock":           "{\n    \"packages\": [\n        {\n            \"name\": \"acme/core\",\n            \"dist\": {\n                \"type\": \"path\",\n                \"url\": \"libs/core\"\n            }\n        }\n    ]\n}\n",
	"libs/core/composer.json": coreManifest,
	"apps/cli/composer.json":  "{\n    \"name\": \"acme/cli\",\n    \"type\": \"project\",\n    \"bin\": [\"bin/cli\"],\n    \"extra\": {\n        \"smith\": {\n            \"kind\": \"app\"\n        }\n    }\n}\n",
	"apps/cli/bin/cli":        "#!/usr/bin/env php\n<?php\nrequire $_composer_autoload_path ?? __DIR__ . '/../../../vendor/autoload.php';\n",
}

func TestLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "core", Root: "libs/core", Language: "php"},
		To:      protocol.ProjectPackage, NewRoot: "packages/core",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	if !strings.Contains(got["libs/core/composer.json"], `"kind": "package"`) {
		t.Fatalf("the kind changes:\n%s", got["libs/core/composer.json"])
	}
	root := got["composer.json"]
	if !strings.Contains(root, `"packages-core": {`) || !strings.Contains(root, `"url": "packages/core"`) || strings.Contains(root, "libs/core") || !strings.Contains(root, `"apps-cli"`) {
		t.Fatalf("the root's path repository follows, key and url:\n%s", root)
	}
	if !strings.Contains(got["composer.lock"], `"url": "packages/core"`) {
		t.Fatalf("the lockfile follows:\n%s", got["composer.lock"])
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "smith install") {
		t.Fatalf("the user is told to install again: %v", r.Warnings)
	}
}

func TestLibToPackageAndBackLeavesTheManifestByteForByte(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/core/composer.json": coreManifest})
	project := protocol.Project{Name: "core", Root: "libs/core", Language: "php"}
	for _, to := range []protocol.ProjectKind{protocol.ProjectPackage, protocol.ProjectLib} {
		r, err := Change(dir, protocol.ChangeKindParams{Project: project, To: to, NewRoot: "libs/core"})
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range r.Files {
			if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(e.Path)), []byte(e.Content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "libs/core/composer.json")); string(got) != coreManifest {
		t.Fatalf("round trip changed the manifest:\n%s", got)
	}
}

func TestAppToLibChangesTheTypeAndNamesTheProgram(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "cli", Root: "apps/cli", Language: "php"},
		To:      protocol.ProjectLib, NewRoot: "libs/cli",
	})
	if err != nil {
		t.Fatal(err)
	}
	cli := edited(r)["apps/cli/composer.json"]
	if !strings.Contains(cli, `"type": "library"`) || !strings.Contains(cli, `"kind": "lib"`) {
		t.Fatalf("a lib is a library:\n%s", cli)
	}
	if !strings.Contains(strings.Join(r.Warnings, "\n"), "executables") {
		t.Fatalf("the program is named: %v", r.Warnings)
	}
}

func TestMovingToADifferentDepthFixesTheExecutablesAutoloadPath(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "cli", Root: "apps/cli", Language: "php"},
		To:      protocol.ProjectApp, NewRoot: "apps/group/cli",
	})
	if err == nil {
		bin := edited(r)["apps/cli/bin/cli"]
		if !strings.Contains(bin, "'/../../../../vendor/autoload.php'") {
			t.Fatalf("one more folder up:\n%s", bin)
		}
	} else {
		t.Fatal(err)
	}
}
