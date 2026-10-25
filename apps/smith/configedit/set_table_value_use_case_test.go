package configedit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestSetTableValueChangesOnlyThatLine(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, workspace.FileName)
	before := "[workspace]\nname = \"acme\"\n\n[plugins.external.kiwi]\n# how it starts\ncommand = [\"kiwi\"]\n\n[ci]\nproviders = [\"github\"]\n"
	_ = os.WriteFile(file, []byte(before), 0o644)
	if err := SetTableValue(root, "plugins.external.kiwi", "sha256", strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(file)
	want := "[workspace]\nname = \"acme\"\n\n[plugins.external.kiwi]\n# how it starts\ncommand = [\"kiwi\"]\nsha256 = \"" + strings.Repeat("a", 64) + "\"\n\n[ci]\nproviders = [\"github\"]\n"
	if string(raw) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", raw, want)
	}
	if err := SetTableValue(root, "plugins.external.kiwi", "sha256", "short"); err == nil {
		t.Fatal("an invalid pin was written")
	}
}

func TestQuoteKey(t *testing.T) {
	if QuoteKey("my-plugin_2") != "my-plugin_2" || QuoteKey("a.b") != `"a.b"` {
		t.Fatal("QuoteKey")
	}
}
