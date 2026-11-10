package main

import (
	"strings"
	"testing"
)

func TestPropertiesListsThePlaceholderOrderAndEscapes(t *testing.T) {
	got := Properties(fixture(t, "jetbrains"), "en")
	for _, want := range []string{
		"# {0}=name {1}=project\ncmd.remove.title=Remove {0} from {1}\n",
		"# {0}=count\ncmd.items.one={0} item\n",
		"# {0}=count\ncmd.items.other={0} items\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "cmd.add.title") {
		t.Error("a key the surface does not claim must not be written")
	}
}

func TestPropertiesOfALocaleHoldsOwnKeysInEnglishPositionsAsUnicodeEscapes(t *testing.T) {
	got := Properties(fixture(t, "jetbrains"), "pt-BR")
	for _, want := range []string{
		"# {0}=name {1}=project\ncmd.remove.title=Remover {1} de {0} \\u2014 a\\u00e7\\u00e3o\n",
		"cmd.items.other={0} itens\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "cmd.items.one") {
		t.Error("pt-BR has no own text for cmd.items.one: the bundle falls back to English")
	}
}

func TestAnApostropheIsDoubledOnlyWhereThereArePlaceholders(t *testing.T) {
	r := Resolved{Keys: []string{"a", "b"}, Texts: map[string]map[string]Entry{"en": {
		"a": {Text: "Don't {name}", Own: true},
		"b": {Text: "Don't", Own: true},
	}}}
	got := Properties(r, "en")
	if !strings.Contains(got, "a=Don''t {0}\n") || !strings.Contains(got, "b=Don't\n") {
		t.Fatalf("got\n%s", got)
	}
}

func TestPropertiesFileNameUsesAnUnderscoreLocale(t *testing.T) {
	for locale, want := range map[string]string{"en": "messages.properties", "pt-BR": "messages_pt_BR.properties"} {
		if got := PropertiesFileName(locale); got != want {
			t.Errorf("PropertiesFileName(%q) = %q, want %q", locale, got, want)
		}
	}
}

func TestPropertiesEscapeHandlesKeysValuesAndSurrogates(t *testing.T) {
	backslash := "\\"
	for _, c := range []struct {
		in, want string
		key      bool
	}{
		{"a b:c=d", `a\ b\:c\=d`, true},
		{" lead", `\ lead`, false},
		{"mid space", "mid space", false},
		{`back\slash`, `back\\slash`, false},
		{"line\nbreak", `line\nbreak`, false},
		// An emoji is a surrogate pair: two UTF-16 escapes.
		{string(rune(0x1F600)), backslash + "ud83d" + backslash + "ude00", false},
	} {
		if got := propertiesEscape(c.in, c.key); got != c.want {
			t.Errorf("propertiesEscape(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
