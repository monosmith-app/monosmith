package discovery

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func declaredWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	files := map[string]string{
		"Package.swift":                   "// swift-tools-version: 5.9\nlet package = Package(name: \"kit\", targets: [.target(name: \"Core\"), .executableTarget(name: \"Cli\")])\n",
		"Sources/Core/Core.swift":         "import Foundation\npublic func hi() {}\n",
		"Sources/Cli/main.swift":          "import Core\n@testable import Foundation\nhi()\n",
		"Tests/CoreTests/CoreTests.swift": "import XCTest\n",
		"empty/Package.swift":             "",
	}
	for name, body := range files {
		full := filepath.Join(ws, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func TestDeclaredFoldersAreTargetsOfThePackageAboveThem(t *testing.T) {
	ws := declaredWorkspace(t)
	res, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "cli", Root: "Sources/Cli"}, {Name: "core", Root: "Sources/Core"}})
	if err != nil || len(res.Projects) != 2 {
		t.Fatal(err, res.Projects)
	}
	cli, core := res.Projects[0], res.Projects[1]
	if cli.Kind != protocol.ProjectApp || !slices.Equal(cli.Exports, []string{"Cli"}) || !slices.Equal(cli.Imports, []string{"Core", "Foundation"}) {
		t.Errorf("cli %+v", cli)
	}
	if core.Kind != protocol.ProjectLib || !slices.Equal(core.Exports, []string{"Core"}) {
		t.Errorf("core %+v", core)
	}
	if got := strings.Join(core.Tasks["build"].Command, " "); got != "swift build --package-path ../.. --target Core --scratch-path ../../dist/Sources/Core" {
		t.Errorf("build %q", got)
	}
	if got := strings.Join(core.Tasks["test"].Command, " "); got != `swift test --package-path ../.. --filter ^CoreTests\.` {
		t.Errorf("test %q", got)
	}
	if !slices.Contains(core.Tasks["build"].Inputs, "{workspaceRoot}/Package.swift") || !slices.Contains(core.Tasks["test"].Inputs, "{workspaceRoot}/Tests/CoreTests/**/*.swift") {
		t.Errorf("inputs %v", core.Tasks)
	}
}

func TestDeclaredFolderIsRefusedWithTheFix(t *testing.T) {
	ws := declaredWorkspace(t)
	bare := t.TempDir()
	if err := os.MkdirAll(filepath.Join(bare, "Sources", "A"), 0o755); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		ws   string
		d    protocol.DeclaredProject
		want string
	}{
		{ws, protocol.DeclaredProject{Name: "x", Root: "Sources/Nope"}, "the folder does not exist"},
		{ws, protocol.DeclaredProject{Name: "own", Root: "empty"}, "Package.swift of its own"},
		{ws, protocol.DeclaredProject{Name: "core", Root: "Sources/Core", Kind: protocol.ProjectPackage}, "one version"},
		{bare, protocol.DeclaredProject{Name: "a", Root: "Sources/A"}, "no Package.swift is above"},
	}
	for _, c := range cases {
		_, err := DiscoverWithDeclared(c.ws, nil, []protocol.DeclaredProject{c.d})
		if err == nil || !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), "[projects."+c.d.Name+"]") {
			t.Errorf("%s: %v, want %q", c.d.Root, err, c.want)
		}
	}
}
