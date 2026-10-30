package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-objc/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	info := protocol.WorkspaceInfo{Name: "acme"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "string-kit", Directory: "libs/string-kit", Workspace: info, Stack: "apple"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core", Workspace: info}); err != nil {
		t.Fatal(err)
	}
	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	spec := read("libs/string-kit/StringKit.podspec")
	if !strings.HasPrefix(spec, "# smithKind: lib\n# smithStack: apple\nPod::Spec.new do |s|\n  s.name         = 'StringKit'\n") || !strings.Contains(spec, "'StringKit/Classes/**/*.{h,m}'") {
		t.Errorf("podspec:\n%s", spec)
	}
	if !strings.Contains(read("libs/string-kit/StringKit/Classes/StringKit.h"), "@interface StringKit : NSObject") || !strings.Contains(read("libs/string-kit/StringKit/Classes/StringKit.m"), `[NSString stringWithFormat:@"Hello, %@!", who]`) {
		t.Error("a class and its header")
	}
	if strings.Contains(read("packages/core/Core.podspec"), "smithStack") {
		t.Error("no stack, no stack comment")
	}
	found, err := discovery.Discover(ws, []string{"libs/string-kit/StringKit.podspec", "packages/core/Core.podspec"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatal(err, found.Projects)
	}
	lib, pkg := found.Projects[0], found.Projects[1]
	if lib.Kind != protocol.ProjectLib || lib.Version != "0.0.0" || lib.Stack != "apple" || lib.Exports[0] != "StringKit" || pkg.Kind != protocol.ProjectPackage {
		t.Fatalf("%+v %+v", lib, pkg)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "x", Directory: "apps/x"}); err == nil {
		t.Error("an app is refused")
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "core", Directory: "packages/core"}); err == nil {
		t.Error("an existing directory is refused")
	}
}

func TestModuleName(t *testing.T) {
	for in, want := range map[string]string{"core": "Core", "string-kit": "StringKit", "2fa": "M2fa", "": "M"} {
		if got := moduleName(in); got != want {
			t.Errorf("moduleName(%q) = %q, want %q", in, got, want)
		}
	}
}
