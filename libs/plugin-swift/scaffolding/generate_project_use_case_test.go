package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-swift/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "sales-report", Directory: "apps/sales-report", Stack: "apple"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "strings", Directory: "libs/strings"}); err != nil {
		t.Fatal(err)
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core"})
	if err != nil || !strings.Contains(res.Notes[0], "tag_prefix") {
		t.Fatalf("a package says how SwiftPM finds its tags: %+v %v", res, err)
	}
	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	app := read("apps/sales-report/Package.swift")
	if !strings.Contains(app, "// smithStack: apple") || !strings.Contains(app, `.executableTarget(name: "SalesReport")`) || strings.Contains(app, "products:") {
		t.Errorf("app:\n%s", app)
	}
	if !strings.Contains(read("apps/sales-report/Sources/SalesReport/main.swift"), "print(greeting())") || !strings.Contains(read("apps/sales-report/Tests/SalesReportTests/SalesReportTests.swift"), "@testable import SalesReport") {
		t.Error("an app has an entry point and a test of it")
	}
	lib := read("libs/strings/Package.swift")
	if !strings.HasPrefix(lib, "// swift-tools-version: 5.9\n// smithKind: lib\n// smithVersion: 0.0.0\n") || !strings.Contains(lib, `.library(name: "Strings"`) || strings.Contains(lib, "smithStack") {
		t.Errorf("lib:\n%s", lib)
	}
	if !strings.Contains(read("libs/strings/Sources/Strings/Strings.swift"), "public func greet") || read("libs/strings/.gitignore") != ".build/\n.swiftpm/\n" {
		t.Error("a lib has a public function and ignores .build")
	}

	found, err := discovery.Discover(ws, []string{"apps/sales-report/Package.swift", "libs/strings/Package.swift", "packages/core/Package.swift"})
	if err != nil || len(found.Projects) != 3 {
		t.Fatal(err, found.Projects)
	}
	app0, lib0, pkg0 := found.Projects[0], found.Projects[1], found.Projects[2]
	if app0.Kind != protocol.ProjectApp || app0.Version != "0.0.0" || app0.Stack != "apple" || lib0.Kind != protocol.ProjectLib || pkg0.Kind != protocol.ProjectPackage {
		t.Fatalf("%+v %+v %+v", app0, lib0, pkg0)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "strings", Directory: "libs/strings"}); err == nil {
		t.Error("an existing directory is refused")
	}
}

func TestModuleName(t *testing.T) {
	for in, want := range map[string]string{"report": "Report", "sales-report": "SalesReport", "2fa": "M2fa", "": "M", "my_lib.v2": "MyLibV2"} {
		if got := moduleName(in); got != want {
			t.Errorf("moduleName(%q) = %q, want %q", in, got, want)
		}
	}
}
