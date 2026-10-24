package manifestfixes

import (
	"strings"
	"testing"
)

func TestInstallMissingPackagesFixInstallsOnlyWhatIsMissing(t *testing.T) {
	root := workspaceWith(t, map[string]string{
		"package.json":            `{"name": "w", "workspaces": ["packages/*"], "devDependencies": {"eslint": "^10.0.0"}}`,
		"packages/a/package.json": `{"name": "a", "devDependencies": {"@eslint/js": "^10.0.1"}}`,
	})
	fix := InstallMissingPackagesFix(root, []string{"eslint", "@eslint/js", "globals"})
	if fix == nil || fix.Safe || len(fix.Steps) != 2 {
		t.Fatalf("%+v", fix)
	}
	if got := strings.Join(fix.Steps[0].Argv, " "); got != "npm install --no-audit --no-fund" {
		t.Fatalf("a declared package only needs npm install: %s", got)
	}
	if got := strings.Join(fix.Steps[1].Argv, " "); got != "npm install -D --no-audit --no-fund @eslint/js@^10.0.1 globals" {
		t.Fatalf("an undeclared one is added, at the range the workspace uses: %s", got)
	}
	if InstallMissingPackagesFix(root, nil) != nil {
		t.Fatal("nothing missing, no fix")
	}
}
