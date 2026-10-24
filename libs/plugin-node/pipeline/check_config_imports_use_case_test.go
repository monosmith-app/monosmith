package pipeline

import (
	"strings"
	"testing"
)

func TestCheckConfigImportsOffersToInstallExactlyTheMissingPackages(t *testing.T) {
	ws := t.TempDir()
	writeFile(t, ws, "package.json", `{"devDependencies":{"@eslint/js":"^10.0.1"}}`)
	writeFile(t, ws, "eslint.config.mjs", "import js from '@eslint/js'\nimport globals from 'globals'\nimport eslint from 'eslint'\n")
	writeFile(t, ws, "node_modules/eslint/package.json", `{"name":"eslint"}`)

	f, ok := CheckConfigImports(ws)
	if !ok || f.OK {
		t.Fatalf("finding %+v", f)
	}
	if f.Fix == nil || f.Fix.Safe || len(f.Fix.Steps) != 2 {
		t.Fatalf("fix %+v", f.Fix)
	}
	if got := strings.Join(f.Fix.Steps[1].Argv, " "); !strings.HasSuffix(got, "-D --no-audit --no-fund globals") {
		t.Errorf("only globals is added; @eslint/js is declared and needs the plain install: %s", got)
	}

	writeFile(t, ws, "node_modules/globals/package.json", `{"name":"globals"}`)
	writeFile(t, ws, "node_modules/@eslint/js/package.json", `{"name":"@eslint/js"}`)
	if f, _ := CheckConfigImports(ws); !f.OK || f.Fix != nil {
		t.Errorf("installed: %+v", f)
	}
}
