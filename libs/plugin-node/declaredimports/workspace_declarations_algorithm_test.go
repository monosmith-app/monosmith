package declaredimports

import "testing"

func TestATestMayImportWhatASiblingDeclaresButShippedCodeMayNot(t *testing.T) {
	root := t.TempDir()
	write(t, root, "package.json", `{"private": true, "workspaces": ["packages/*"]}`)
	write(t, root, "packages/a/package.json", `{"name": "@acme/a", "version": "1.0.0"}`)
	write(t, root, "packages/a/src/a.ts", `import { x } from 'only-a-sibling-declares'`)
	write(t, root, "packages/a/src/a.test.ts", `import { x } from 'only-a-sibling-declares'
import { y } from 'nobody-declares'`)
	write(t, root, "packages/b/package.json", `{"name": "@acme/b", "devDependencies": {"only-a-sibling-declares": "^1.0.0"}}`)
	got, err := Check(root, "packages/a")
	if err != nil {
		t.Fatal(err)
	}
	want := []UndeclaredImport{
		{File: "src/a.test.ts", Package: "nobody-declares"},
		{File: "src/a.ts", Package: "only-a-sibling-declares"},
	}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %+v\nwant %+v", got, want)
	}
}

func TestWorkspacePatternsReadsBothFormsOfWorkspaces(t *testing.T) {
	for raw, want := range map[string]int{`["a/*", "b/*"]`: 2, `{"packages": ["a/*"]}`: 1, `null`: 0, ``: 0} {
		if got := len(workspacePatterns([]byte(raw))); got != want {
			t.Errorf("%q: %d patterns, want %d", raw, got, want)
		}
	}
}
