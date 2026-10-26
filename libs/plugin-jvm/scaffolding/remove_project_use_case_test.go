package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestRemoveTakesAMavenModuleOutOfTheRootPom(t *testing.T) {
	ws := t.TempDir()
	pom := "<project>\n  <modules>\n    <module>apps/web</module>\n    <module>libs/core</module>\n  </modules>\n</project>\n"
	if err := os.WriteFile(filepath.Join(ws, "pom.xml"), []byte(pom), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Remove(ws, protocol.RemoveParams{Project: protocol.Project{Name: "web", Root: "apps/web"}})
	if err != nil || strings.Join(res.Files, ",") != "pom.xml" {
		t.Fatalf("%v %v", err, res.Files)
	}
	got, _ := os.ReadFile(filepath.Join(ws, "pom.xml"))
	if strings.Contains(string(got), "apps/web") || !strings.Contains(string(got), "<module>libs/core</module>") {
		t.Errorf("pom:\n%s", got)
	}
}

func TestRemoveTakesAGradleIncludeOutOfTheSettings(t *testing.T) {
	ws := t.TempDir()
	settingsText := "rootProject.name = \"acme\"\n\ninclude(\":apps:web\")\ninclude(\":libs:core\")\n"
	if err := os.WriteFile(filepath.Join(ws, "settings.gradle.kts"), []byte(settingsText), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Remove(ws, protocol.RemoveParams{Project: protocol.Project{Name: "web", Root: "apps/web"}})
	if err != nil || strings.Join(res.Files, ",") != "settings.gradle.kts" {
		t.Fatalf("%v %v", err, res.Files)
	}
	got, _ := os.ReadFile(filepath.Join(ws, "settings.gradle.kts"))
	if strings.Contains(string(got), ":apps:web") || !strings.Contains(string(got), `include(":libs:core")`) {
		t.Errorf("settings:\n%s", got)
	}
}

func TestRemoveDryRunReportsTheFileAndChangesNothing(t *testing.T) {
	ws := t.TempDir()
	pom := "<project>\n  <modules>\n    <module>apps/web</module>\n  </modules>\n</project>\n"
	if err := os.WriteFile(filepath.Join(ws, "pom.xml"), []byte(pom), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Remove(ws, protocol.RemoveParams{DryRun: true, Project: protocol.Project{Name: "web", Root: "apps/web"}})
	if err != nil || len(res.Files) != 1 {
		t.Fatalf("%v %v", err, res.Files)
	}
	if got, _ := os.ReadFile(filepath.Join(ws, "pom.xml")); string(got) != pom {
		t.Errorf("a dry run changed the pom:\n%s", got)
	}
}

func TestRemoveWithNoBuildFileIsNotAnError(t *testing.T) {
	if res, err := Remove(t.TempDir(), protocol.RemoveParams{Project: protocol.Project{Root: "apps/web"}}); err != nil || len(res.Files) != 0 {
		t.Fatalf("%v %v", err, res.Files)
	}
}
