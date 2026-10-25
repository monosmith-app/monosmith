package convertchoices

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
)

func TestAskTakesTheSuggestedOptionOnAnEmptyAnswer(t *testing.T) {
	def := Choices{Level: LevelFull, Lint: Smith, Test: Jest, Build: Smith, CI: Smith}
	qs := Questions(legacyinventory.Usage{ESLint: true, Jest: true}, def)
	var out bytes.Buffer
	// lint: empty (suggested), test: 3 (keep), build: given, ci: 2 (keep).
	in := bufio.NewReader(strings.NewReader("\n3\n2\n"))
	got, err := Ask(in, &out, qs, def, Given{Build: true})
	if err != nil {
		t.Fatal(err)
	}
	if got.Lint != Smith || got.Test != Keep || got.Build != Smith || got.CI != Keep {
		t.Fatalf("choices = %+v", got)
	}
	if !strings.Contains(out.String(), "1) smith's ESLint config") || !strings.Contains(out.String(), "(suggested)") || strings.Contains(out.String(), "Build with") {
		t.Fatalf("output:\n%s", out.String())
	}
}

func TestAskRejectsAnAnswerThatIsNotAnOption(t *testing.T) {
	def := Defaults(legacyinventory.Usage{ESLint: true, Jest: true})
	_, err := Ask(bufio.NewReader(strings.NewReader("9\n")), &bytes.Buffer{}, Questions(legacyinventory.Usage{ESLint: true, Jest: true}, def), def, Given{})
	if err == nil {
		t.Fatal("want an error")
	}
}
