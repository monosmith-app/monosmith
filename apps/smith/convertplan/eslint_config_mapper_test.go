package convertplan

import (
	"os"
	"strings"
	"testing"
)

func TestRewriteNxESLintRoot(t *testing.T) {
	src, _ := os.ReadFile("testdata/nx-root-eslint.config.mjs")
	got, ok := RewriteNxESLint(string(src))
	if !ok {
		t.Fatalf("not rewritten:\n%s", got)
	}
	for _, want := range []string{"import js from '@eslint/js';", "import tseslint from 'typescript-eslint';", "  js.configs.recommended,", "  ...tseslint.configs.recommended,", "ignores: ['**/dist'", "// Override or add rules here"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "nx") || strings.Contains(got, "enforce-module-boundaries") {
		t.Fatalf("Nx left in:\n%s", got)
	}
}

func TestRewriteNxESLintProject(t *testing.T) {
	src, _ := os.ReadFile("testdata/nx-project-eslint.config.mjs")
	got, ok := RewriteNxESLint(string(src))
	if !ok {
		t.Fatalf("not rewritten:\n%s", got)
	}
	want := `import baseConfig from '../../eslint.config.mjs';

export default [
  ...baseConfig,
  {
    ignores: ['**/out-tsc'],
  },
];
`
	if strings.ReplaceAll(got, "\r\n", "\n") != want {
		t.Fatalf("got:\n%s", got)
	}
}

func TestRewriteNxESLintRefusesUnknownPresets(t *testing.T) {
	src := "import nx from '@nx/eslint-plugin';\nexport default [...nx.configs['flat/react']];\n"
	if _, ok := RewriteNxESLint(src); ok {
		t.Fatal("a preset with no plain equivalent must stay manual")
	}
}

// A dependency check on package.json files is dropped with the comment that
// explains it, and a comment that only names Nx does not stop the rewrite.
func TestRewriteNxESLintDropsTheDependencyChecksBlockWithItsComment(t *testing.T) {
	src := "// Built on @nx/eslint-plugin's checks; see below.\nimport base from './base.mjs'\n\nexport default [\n  base,\n  {\n    // @nx/dependency-checks cannot see a @types/* package: they are not imported.\n    // These are not unused.\n    name:  'local/types',\n    files: ['packages/*/package.json'],\n    rules: {\n      '@nx/dependency-checks': ['error', { ignoredDependencies: ['@types/he'] }],\n    },\n  },\n  { name: 'local/keep', rules: { 'no-undef': 'off' } },\n]\n"
	out, ok := RewriteNxESLint(src)
	if !ok {
		t.Fatalf("not rewritten:\n%s", src)
	}
	if strings.Contains(out, "dependency-checks cannot see") || strings.Contains(out, "@types/he") || strings.Contains(out, "local/types") {
		t.Errorf("the block or its comment stayed:\n%s", out)
	}
	if !strings.Contains(out, "local/keep") || !strings.Contains(out, "@nx/eslint-plugin's checks; see below") {
		t.Errorf("what is not about the check was lost:\n%s", out)
	}
}
