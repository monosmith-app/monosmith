package main

import (
	"encoding/json"
	"testing"
)

func decode(t *testing.T, text string) map[string]string {
	t.Helper()
	var out map[string]string
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, text)
	}
	return out
}

func TestNlsJSONHoldsEveryKeyInEnglishWithPositionalPlaceholders(t *testing.T) {
	got := decode(t, NlsJSON(fixture(t, "vscode"), "en"))
	want := map[string]string{
		"cmd.add.title":    "Add a package",
		"cmd.remove.title": "Remove {0} from {1}",
		"vscode.quote":     "Don't remove {0} \"now\"",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestNlsJSONOfALocaleHoldsOnlyItsOwnKeysInTheEnglishPositions(t *testing.T) {
	got := decode(t, NlsJSON(fixture(t, "vscode"), "pt-BR"))
	if len(got) != 1 || got["cmd.remove.title"] != "Remover {1} de {0} — ação" {
		t.Fatalf("got %v", got)
	}
}

func TestNlsFileNameLowerCasesTheLocale(t *testing.T) {
	for locale, want := range map[string]string{"en": "package.nls.json", "pt-BR": "package.nls.pt-br.json"} {
		if got := NlsFileName(locale); got != want {
			t.Errorf("NlsFileName(%q) = %q, want %q", locale, got, want)
		}
	}
}

func TestNlsJSONOfNoKeysIsAnEmptyObject(t *testing.T) {
	if got := NlsJSON(fixture(t, "vscode"), "xx"); got != "{}\n" {
		t.Fatalf("got %q", got)
	}
}
