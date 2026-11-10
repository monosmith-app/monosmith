package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAFolderOfTheModuleNoProjectHoldsIsReportedToTheProjectThatImportsIt(t *testing.T) {
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "go.mod"), []byte("module example.com/mvd\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "apps", "mvd"), 0o755); err != nil {
		t.Fatal(err)
	}
	projects := []protocol.Project{
		{Name: "mvd", Root: "apps/mvd", Exports: []string{"example.com/mvd/apps/mvd"}, Imports: []string{"example.com/mvd/libs/core/greeting", "example.com/mvd/internal/store", "fmt", "github.com/x/y"}},
		{Name: "core", Root: "libs/core", Exports: []string{"example.com/mvd/libs/core"}},
	}
	got := CheckUncoveredImports(ws, projects)
	if len(got) != 1 || !got[0].Warning || !strings.Contains(got[0].Detail, "apps/mvd imports internal/store of its module, which no project holds") || strings.Contains(got[0].Detail, "greeting") {
		t.Fatalf("one warning for the folder no project holds, none for a covered sub-package or another module: %+v", got)
	}
	if !strings.Contains(got[0].Remedy, "kind = \"lib\"") {
		t.Errorf("the remedy says how to declare it: %q", got[0].Remedy)
	}
	projects = append(projects, protocol.Project{Name: "store", Root: "internal/store", Exports: []string{"example.com/mvd/internal/store"}})
	if again := CheckUncoveredImports(ws, projects); len(again) != 0 {
		t.Errorf("declared, nothing to report: %+v", again)
	}
}
