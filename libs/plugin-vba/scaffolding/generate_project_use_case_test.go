package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-vba/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "sales-report", Directory: "apps/sales-report", Stack: "office"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "strings", Directory: "libs/strings"}); err != nil {
		t.Fatal(err)
	}
	app, _ := os.ReadFile(filepath.Join(ws, "apps", "sales-report", "main.bas"))
	lib, _ := os.ReadFile(filepath.Join(ws, "libs", "strings", "main.bas"))
	if !strings.HasPrefix(string(app), "Attribute VB_Name = \"sales_report\"\r\n' smithKind: app\r\n' smithStack: office\r\n") || !strings.Contains(string(app), "Public Sub Main()") {
		t.Errorf("app: %q", app)
	}
	if !strings.Contains(string(lib), "Public Function Greet") || strings.Contains(string(lib), "smithStack") || strings.Count(string(lib), "\n") != strings.Count(string(lib), "\r\n") {
		t.Errorf("lib: %q", lib)
	}
	found, err := discovery.Discover(ws, []string{"apps/sales-report/main.bas", "libs/strings/main.bas"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatal(err, found.Projects)
	}
	if p := found.Projects[0]; p.Kind != protocol.ProjectApp || p.Version != "0.0.0" || p.Stack != "office" || found.Projects[1].Kind != protocol.ProjectLib {
		t.Fatalf("%+v", found.Projects)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "x", Directory: "packages/x"}); err == nil {
		t.Error("a package is refused")
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "strings", Directory: "libs/strings"}); err == nil {
		t.Error("an existing directory is refused")
	}
}

func TestModuleName(t *testing.T) {
	for in, want := range map[string]string{"report": "report", "sales-report": "sales_report", "2fa": "M2fa", "": "M", "_x": "M_x"} {
		if got := moduleName(in); got != want {
			t.Errorf("moduleName(%q) = %q, want %q", in, got, want)
		}
	}
}
