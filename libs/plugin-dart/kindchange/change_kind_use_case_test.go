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

const corePubspec = `name: acme_core
description: core
version: 0.0.0
publish_to: none
resolution: workspace

environment:
  sdk: ^3.7.0

dependencies:
  acme_util:
    path: ../util

smith:
  kind: lib
`

var workspace = map[string]string{
	"pubspec.yaml":                    "name: acme_workspace\npublish_to: none\n\nworkspace:\n  - libs/core\n  - libs/util\n  - apps/cli\n",
	"pubspec.lock":                    "packages:\n  acme_core:\n    source: path\n    description:\n      path: libs/core\n",
	"libs/core/pubspec.yaml":          corePubspec,
	"libs/core/analysis_options.yaml": "include: ../../analysis_options.yaml\n",
	"libs/util/pubspec.yaml":          "name: acme_util\nsmith:\n  kind: lib\n",
	"apps/cli/pubspec.yaml":           "name: acme_cli\ndependencies:\n  acme_core:\n    path: ../../libs/core\nsmith:\n  kind: app\n",
	"apps/cli/bin/main.dart":          "void main() {}\n",
}

func projects() []protocol.Project {
	return []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "dart"},
		{Name: "util", Root: "libs/util", Language: "dart"},
		{Name: "cli", Root: "apps/cli", Language: "dart"},
	}
}

func TestLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects()[0], To: protocol.ProjectPackage, NewRoot: "packages/core", Projects: projects()})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	core := got["libs/core/pubspec.yaml"]
	if strings.Contains(core, "publish_to") || !strings.Contains(core, "kind: package") {
		t.Fatalf("the publish guard goes and the kind changes:\n%s", core)
	}
	if !strings.Contains(core, "path: ../../libs/util") {
		t.Fatalf("its own path dependency still reaches util:\n%s", core)
	}
	if got["libs/core/analysis_options.yaml"] != "include: ../../analysis_options.yaml\n" {
		if _, touched := got["libs/core/analysis_options.yaml"]; touched {
			t.Fatalf("a path that stays valid is kept: %q", got["libs/core/analysis_options.yaml"])
		}
	}
	root := got["pubspec.yaml"]
	if !strings.Contains(root, "  - packages/core\n") || !strings.Contains(root, "  - libs/util\n") || strings.Contains(root, "libs/core") {
		t.Fatalf("the workspace lists the new directory:\n%s", root)
	}
	if !strings.Contains(got["apps/cli/pubspec.yaml"], "path: ../../packages/core") {
		t.Fatalf("dependents follow:\n%s", got["apps/cli/pubspec.yaml"])
	}
	if !strings.Contains(got["pubspec.lock"], "path: packages/core") {
		t.Fatalf("the lockfile follows:\n%s", got["pubspec.lock"])
	}
	joined := strings.Join(r.Warnings, "\n")
	if !strings.Contains(joined, "repository") || !strings.Contains(joined, "smith install") {
		t.Fatalf("pub.dev's needs and the install are named: %v", r.Warnings)
	}
}

func TestLibToPackageAndBackLeavesThePubspecByteForByte(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/core/pubspec.yaml": corePubspec})
	project := protocol.Project{Name: "core", Root: "libs/core", Language: "dart"}
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
	if got, _ := os.ReadFile(filepath.Join(dir, "libs/core/pubspec.yaml")); string(got) != corePubspec {
		t.Fatalf("round trip changed the pubspec:\n%s", got)
	}
}

func TestAppToLibGetsTheGuardAndNamesTheProgram(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects()[2], To: protocol.ProjectLib, NewRoot: "libs/cli", Projects: projects()})
	if err != nil {
		t.Fatal(err)
	}
	cli := edited(r)["apps/cli/pubspec.yaml"]
	if !strings.Contains(cli, "kind: lib") || !strings.Contains(cli, "publish_to: none") {
		t.Fatalf("a lib is not published:\n%s", cli)
	}
	if !strings.Contains(strings.Join(r.Warnings, "\n"), "program") {
		t.Fatalf("the program is named: %v", r.Warnings)
	}
}

func TestMissingSmithSectionIsAdded(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/x/pubspec.yaml": "name: x\nversion: 0.0.0\n"})
	r, err := Change(dir, protocol.ChangeKindParams{Project: protocol.Project{Name: "x", Root: "libs/x", Language: "dart"}, To: protocol.ProjectLib, NewRoot: "libs/x"})
	if err != nil {
		t.Fatal(err)
	}
	if got := edited(r)["libs/x/pubspec.yaml"]; got != "name: x\nversion: 0.0.0\npublish_to: none\n\nsmith:\n  kind: lib\n" {
		t.Fatalf("got %q", got)
	}
}
