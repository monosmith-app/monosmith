package scaffold

import (
	"bytes"
	"strings"
	"testing"
)

func TestPickLanguageSaysWhenItChoseForYou(t *testing.T) {
	lang, note, err := pickLanguage("", "site", []string{"go"}, []string{"bicep", "go", "node"}, nil, nil, false)
	if err != nil || lang != "go" {
		t.Fatalf("%q %v", lang, err)
	}
	if !strings.Contains(note, "No --lang: using go, the only language in this workspace (pass --lang with another to choose: node, ...; `smith kinds`") || strings.Contains(note, "bicep") {
		t.Fatalf("%q", note)
	}
}

func TestPickLanguageAsksATerminal(t *testing.T) {
	var out bytes.Buffer
	lang, note, err := pickLanguage("", "site", []string{"go"}, []string{"go", "node"}, strings.NewReader("2\n"), &out, true)
	if err != nil || lang != "node" || note != "" {
		t.Fatalf("%q %q %v", lang, note, err)
	}
	if !strings.Contains(out.String(), "Which language is \"site\"?") || !strings.Contains(out.String(), "1) go  (default") {
		t.Fatalf("%s", out.String())
	}
	if lang, _, _ := pickLanguage("", "site", []string{"go"}, []string{"go", "node"}, strings.NewReader("\n"), &out, true); lang != "go" {
		t.Fatalf("an empty answer takes the default, got %q", lang)
	}
}

func TestPickLanguageKeepsAnExplicitLanguageAndRefusesToGuess(t *testing.T) {
	if lang, note, err := pickLanguage("python", "x", []string{"go"}, []string{"go", "python"}, nil, nil, false); lang != "python" || note != "" || err != nil {
		t.Fatalf("%q %q %v", lang, note, err)
	}
	if _, _, err := pickLanguage("", "x", []string{"go", "node"}, []string{"go", "node"}, nil, nil, false); err == nil || !strings.Contains(err.Error(), "--lang (go, node)") {
		t.Fatalf("%v", err)
	}
}
