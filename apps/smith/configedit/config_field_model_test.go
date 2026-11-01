package configedit

import (
	"strings"
	"testing"
)

func TestAStringThatTakesAFewWordsIsEditedWithItsChoices(t *testing.T) {
	for _, path := range []string{"ci.attest", "release.api_check"} {
		f, err := Find(path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if f.Type != "string" || len(f.Enum) == 0 {
			t.Fatalf("%s: type %q, choices %v", path, f.Type, f.Enum)
		}
		word := f.Enum[0]
		if got, err := f.Encode([]string{word}); err != nil || got != `"`+word+`"` {
			t.Errorf("%s: %q, %v", path, got, err)
		}
		if _, err := f.Encode([]string{"maybe"}); err == nil || !strings.Contains(err.Error(), "takes "+strings.Join(f.Enum, " or ")) {
			t.Errorf("%s: a word outside the choices must name them, got %v", path, err)
		}
	}
}

func TestAListOfTablesIsNotASettingSmithConfigEdits(t *testing.T) {
	if _, err := Find("release.channels"); err == nil {
		t.Error("release.channels is edited in smith.toml: a list of tables")
	}
	fields, _ := Fields()
	for _, f := range fields {
		if f.Type == "" {
			t.Errorf("%s has no type, so smith config could not edit it", f.Path)
		}
	}
}
