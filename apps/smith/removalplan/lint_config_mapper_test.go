package removalplan

import (
	"strings"
	"testing"
)

const sharedConfig = `import mnci from '@mnci/eslint-config'

// ...mnci({ workspaceRoot: import.meta.dirname }) is what the comment says.
export default [
  ...mnci({ workspaceRoot: import.meta.dirname }),
  { name: 'local/rules', rules: { 'no-console': 'off' } },
]
`

func TestSwapLintConfigKeepsTheRulesWrittenBesideTheSharedConfig(t *testing.T) {
	got, how, _, ok := SwapLintConfig(sharedConfig)
	if !ok || how != LintSwapped {
		t.Fatalf("how = %q ok = %v", how, ok)
	}
	for _, want := range []string{"import monosmith from '@monosmith/eslint-config'", "...monosmith({ workspaceRoot: import.meta.dirname })", "'no-console': 'off'"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "mnci(") || strings.Contains(got, "@mnci/") {
		t.Errorf("the old config is still called:\n%s", got)
	}
}

func TestSwapLintConfigReplacesAHandWrittenConfig(t *testing.T) {
	got, how, _, ok := SwapLintConfig("import js from '@eslint/js'\nexport default [js.configs.recommended]\n")
	if !ok || how != LintReplaced || got != SmithLintConfig {
		t.Fatalf("got %q how %q ok %v", got, how, ok)
	}
}

func TestSwapLintConfigRefusesOptionsSmithDoesNotTake(t *testing.T) {
	_, _, why, ok := SwapLintConfig("import mnci from '@mnci/eslint-config'\nexport default [...mnci({ workspaceRoot: x, react: true })]\n")
	if ok || !strings.Contains(why, "options") {
		t.Fatalf("ok = %v why = %q", ok, why)
	}
}

func TestSwapLintConfigReadsWindowsLineEndings(t *testing.T) {
	_, how, _, ok := SwapLintConfig(strings.ReplaceAll(sharedConfig, "\n", "\r\n"))
	if !ok || how != LintSwapped {
		t.Fatalf("how = %q ok = %v", how, ok)
	}
}

func TestSwapLintConfigLeavesSmithsOwnConfigAlone(t *testing.T) {
	got, how, _, ok := SwapLintConfig(SmithLintConfig)
	if !ok || how != "" || got != SmithLintConfig {
		t.Fatalf("got %q how %q ok %v", got, how, ok)
	}
}
