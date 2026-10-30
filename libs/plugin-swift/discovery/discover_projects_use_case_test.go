package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("apps/cli/Package.swift", `// swift-tools-version: 5.9
// smithVersion: 1.2.0
import PackageDescription

let package = Package(
    name: "cli",
    dependencies: [
        .package(path: "../../libs/core"),
        .package(path: "Vendored"),
        .package(url: "https://github.com/apple/swift-argument-parser", from: "1.3.0"),
    ],
    targets: [.executableTarget(name: "cli", dependencies: [])]
)
`)
	write("libs/core/Package.swift", "// swift-tools-version: 5.9\n// smithStack: apple\n// smithKind: package\nlet package = Package(name: \"core\")\n")
	write("libs/plain/Package.swift", "// swift-tools-version: 5.9\nlet package = Package(name: \"plain\")\n")
	write("Package.swift", "")
	res, err := Discover(ws, []string{"Package.swift", "apps/cli/Package.swift", "libs/core/Package.swift", "libs/plain/Package.swift"})
	if err != nil || len(res.Projects) != 3 {
		t.Fatal(err, res.Projects)
	}
	cli, core, plain := res.Projects[0], res.Projects[1], res.Projects[2]
	if cli.Name != "cli" || cli.Kind != protocol.ProjectApp || cli.Version != "1.2.0" || strings.Join(cli.Imports, ",") != "libs/core" || cli.Exports[0] != "apps/cli" {
		t.Errorf("cli %+v", cli)
	}
	if got := strings.Join(cli.Tasks["build"].Command, " "); got != "swift build --scratch-path ../../dist/apps/cli" || cli.Tasks["build"].Outputs[0] != "../../dist/apps/cli" {
		t.Errorf("build %q", got)
	}
	if core.Kind != protocol.ProjectPackage || core.Stack != "apple" {
		t.Errorf("core %+v", core)
	}
	if plain.Kind != protocol.ProjectLib {
		t.Errorf("a package without an executable is a lib: %+v", plain)
	}
	if _, ok := plain.Tasks["test"]; !ok {
		t.Error("every project has a test task")
	}
	if _, ok := plain.Tasks["lint"]; ok {
		t.Error("no lint task")
	}
}

func TestPackageRoot(t *testing.T) {
	for dep, want := range map[string]string{
		"../../libs/core": "libs/core",
		"Vendored":        "apps/cli/Vendored",
		"../../../out":    "",
		"..":              "apps",
		"../..":           "",
	} {
		if got := PackageRoot("apps/cli", dep); got != want {
			t.Errorf("PackageRoot(%q) = %q, want %q", dep, got, want)
		}
	}
}

func TestSetField(t *testing.T) {
	out, ok := SetField("// swift-tools-version: 5.9\nimport PackageDescription\n", "Version", "1.0.0")
	if !ok || out != "// swift-tools-version: 5.9\n// smithVersion: 1.0.0\nimport PackageDescription\n" {
		t.Errorf("a new comment goes under swift-tools-version: %q", out)
	}
	if _, ok := SetField(out, "Version", "1.0.0"); ok {
		t.Error("the same version is no change")
	}
	if out, _ := SetField(out, "Version", "2.0.0"); !strings.Contains(out, "// smithVersion: 2.0.0\n") {
		t.Errorf("%q", out)
	}
	if out, _ := SetField("import X\r\n", "Kind", "lib"); out != "// smithKind: lib\r\nimport X\r\n" {
		t.Errorf("%q", out)
	}
}
