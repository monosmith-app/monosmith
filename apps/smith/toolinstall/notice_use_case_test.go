package toolinstall

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolcatalog"
)

func TestLanguageToolsAreInTheCatalogue(t *testing.T) {
	for lang, names := range languageTools {
		for _, name := range names {
			if _, ok := toolcatalog.Find(name); !ok {
				t.Errorf("%s: %q is not in the catalogue", lang, name)
			}
		}
	}
}

func TestNoticeTextNamesTheCommandAndStaysQuietWhenAllIsWell(t *testing.T) {
	ok := Status{Need: missing("git").Need, State: Found}
	if got := NoticeText([]Status{ok}); got != "" {
		t.Fatalf("nothing is missing, got %q", got)
	}
	old := missing("python")
	old.State, old.Version, old.Wanted = Outdated, "3.9.1", Requirement{Min: "3.11"}
	got := NoticeText([]Status{missing("dart"), old})
	for _, want := range []string{"! the Dart SDK is not installed", "smith tools install dart", "! Python is too old (3.9.1 is installed, 3.11 or newer is wanted)", "smith tools install python"} {
		if !strings.Contains(got, want) {
			t.Errorf("notice lacks %q:\n%s", want, got)
		}
	}
}
