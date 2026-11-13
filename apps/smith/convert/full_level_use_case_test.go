package convert

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFindingsTextListsEachCheckThenSaysTheConversionFinishedWithFindings(t *testing.T) {
	got := findingsText([]doctorfix.Finding{
		{Step: protocol.FixStep{Description: "fix what ESLint can, then lint every project", Argv: []string{"smith", "lint", "--fix", "-p", "lang:node"}}, Output: "▲ ESLint fixed 3 problems; 2 are left in 1 of 1 project\n"},
		{Step: protocol.FixStep{Description: "check the workspace with doctor", Argv: []string{"smith", "doctor"}}, Output: "× app:build writes apps/app/dist-dev\n    → build each project into dist/<project dir>/\n"},
	})
	for _, want := range []string{
		"▲ check the workspace with doctor\n  × app:build writes apps/app/dist-dev\n    → build each project into dist/<project dir>/\n  → check again: smith doctor\n",
		"▲ The conversion finished with findings: the changes stay, and 2 checks found something to fix (listed above).\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("lacks %q:\n%s", want, got)
		}
	}
	if !strings.HasSuffix(got, "(listed above).\n") || strings.Contains(got, "stopped") {
		t.Fatalf("the closing line is the outcome:\n%s", got)
	}
}

func TestCommandForNamesTheShimNotThePlatformBinary(t *testing.T) {
	notFound := func(string) (string, error) { return "", errors.New("not found") }
	platform := "C:\\Users\\me\\AppData\\Roaming\\npm\\node_modules\\monosmith\\node_modules\\@monosmith\\smith-win32-x64\\bin\\smith.exe"
	if got := commandFor(platform, notFound); got != "smith" {
		t.Errorf("an npm platform binary is named %q", got)
	}
	if got := commandFor("/home/me/build/smith", notFound); got != "/home/me/build/smith" {
		t.Errorf("a binary that is not on the path keeps its path, got %q", got)
	}
}
