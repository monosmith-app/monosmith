package palette

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

func TestTheTypedLettersShowInTheAccent(t *testing.T) {
	th := theme.Default()
	got := typed(New(ui), "doc").titleLine(th, ui[2], false, 60)
	if !strings.Contains(got, th.Bold(th.Accent("Doc"))) {
		t.Fatalf("the match is not lit: %q", got)
	}
	if spans := matchSpans("Build release web", "build web"); len(spans) != 2 || spans[1][0] != 14 {
		t.Fatalf("each typed word is lit: %v", spans)
	}
	if matchSpans("Build", "") != nil {
		t.Fatal("nothing typed lights nothing")
	}
}

func TestTheCommandLineIsTheHintUpToItsDescription(t *testing.T) {
	for hint, want := range map[string]string{
		"smith add: an app, lib or package": "smith add",
		"smith run build -p web":            "smith run build -p web",
		"cancel the run":                    "cancel the run",
	} {
		if got := commandLine(hint); got != want {
			t.Errorf("%q gave %q, want %q", hint, got, want)
		}
	}
}

func TestNothingFoundSaysSoAndHowToFindMore(t *testing.T) {
	text := pane.Strip(strings.Join(typed(New(intents), "zzqxj").Lines(theme.Default(), 72, 14), "\n"))
	if !strings.Contains(text, "no command matches") || !strings.Contains(text, "Type fewer letters") {
		t.Fatalf("%s", text)
	}
}
