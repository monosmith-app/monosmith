package main

import (
	"strings"
	"testing"
)

func TestTypeScriptHoldsReferencesTextsAndTheSayHelper(t *testing.T) {
	got := TypeScript(fixture(t, "vscode"))
	for _, want := range []string{
		`"cmd.remove.title": "%cmd.remove.title%",`,
		`export type WordKey = keyof typeof nls`,
		// English text keeps {name}: say fills by name.
		`"cmd.remove.title": "Remove {name} from {project}",`,
		// pt-br is complete: the key it lacks has the English text.
		`"cmd.add.title": "Add a package",`,
		`"cmd.remove.title": "Remover {project} de {name} — ação",`,
		`  "pt-br": {`,
		`export function say(key: WordKey`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "import ") {
		t.Error("the module must not import anything (vitest loads it without vscode)")
	}
}

func TestTypeScriptOfNoKeysIsAValidEmptyModule(t *testing.T) {
	got := TypeScript(Resolved{Texts: map[string]map[string]Entry{"en": {}}})
	for _, want := range []string{"export const nls = {} as const", "const english: Texts = {}", "const translations: Partial<Record<string, Texts>> = {}", "export function say("} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
}
