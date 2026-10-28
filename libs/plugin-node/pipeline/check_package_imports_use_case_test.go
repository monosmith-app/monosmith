package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctorChecksThatAPackageDeclaresItsImports(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("package.json", `{"workspaces":["packages/*"]}`)
	write("packages/ocr/package.json", `{"name":"@acme/ocr"}`)
	write("packages/ocr/src/read.ts", `import type { Page } from '@acme/core'`)
	write("packages/web/package.json", `{"name":"@acme/web","private":true}`)
	write("packages/web/src/main.ts", `import { x } from 'undeclared'`)
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{
		{Name: "ocr", Root: "packages/ocr", Kind: protocol.ProjectPackage},
		{Name: "web", Root: "packages/web", Kind: protocol.ProjectApp},
	}}, "24")
	var found *protocol.Finding
	for i, f := range res.Findings {
		if strings.HasPrefix(f.Check, "packages/web ") {
			t.Errorf("an app is not checked: %+v", f)
		}
		if f.Check == "packages/ocr declares the packages it imports" {
			found = &res.Findings[i]
		}
	}
	if found == nil || found.OK || found.Detail != "src/read.ts imports @acme/core, which package.json does not declare" {
		t.Fatalf("finding %+v", found)
	}
	// No project declares @acme/core and it is not installed: its version is
	// unknown, so smith offers no fix rather than a guess.
	if found.Fix != nil {
		t.Errorf("fix %+v", found.Fix)
	}
	write("packages/core/package.json", `{"name":"@acme/core","version":"2.1.0"}`)
	write("node_modules/@acme/core/package.json", `{"name":"@acme/core","version":"2.1.0"}`)
	f := CheckPackageImports(ws, protocol.Project{Name: "ocr", Root: "packages/ocr", Kind: protocol.ProjectPackage})
	if f.Fix == nil || !strings.Contains(f.Fix.Files[0].Content, `"@acme/core": "^2.1.0"`) || f.Fix.Safe {
		t.Fatalf("the installed version is the fallback, and installing is asked about: %+v", f.Fix)
	}
}

func TestAnUnreadablePackageGetsNoAddDependenciesRemedy(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "packages", "bad"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"workspaces":["packages/*"]}`), 0o644)
	_ = os.WriteFile(filepath.Join(ws, "packages", "bad", "package.json"), []byte(`{not json`), 0o644)
	f := CheckPackageImports(ws, protocol.Project{Name: "bad", Root: "packages/bad", Kind: protocol.ProjectPackage})
	if f.OK || f.Detail == "" || strings.Contains(f.Remedy, "add each to dependencies") {
		t.Fatalf("finding %+v", f)
	}
}
