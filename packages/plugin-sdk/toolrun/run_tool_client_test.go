package toolrun

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// TestMain lets the test binary act as a tool: with TOOLRUN_HELPER set it
// prints a go test result line, a line it coloured itself behind a cursor
// code, and what the environment asked it for.
func TestMain(m *testing.M) {
	if os.Getenv("TOOLRUN_HELPER") != "" {
		fmt.Println("--- FAIL: TestX (0.00s)")
		fmt.Println("\x1b[2K\x1b[32mgreen by the tool\x1b[0m")
		fmt.Println("FORCE_COLOR=" + os.Getenv("FORCE_COLOR") + " NO_COLOR=" + os.Getenv("NO_COLOR"))
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func helperTool(t *testing.T) []string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return []string{self, "-test.run=^$"}
}

func handOn(t *testing.T, colour bool) {
	t.Helper()
	t.Setenv("TOOLRUN_HELPER", "1")
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "")
	t.Setenv(termcolour.ToolColourEnv, map[bool]string{true: "1", false: "0"}[colour])
}

func TestRunAsksForColourAndPaintsWhenTheHostHandsItOn(t *testing.T) {
	handOn(t, true)
	log := captureLog(t)
	if err := Run(t.TempDir(), helperTool(t)); err != nil {
		t.Fatal(err)
	}
	text := log.String()
	for _, want := range []string{
		termcolour.Wrap("--- FAIL: TestX (0.00s)", termcolour.Red) + "\n",
		"\x1b[32mgreen by the tool\x1b[0m\n",
		"FORCE_COLOR=1 NO_COLOR=\n",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
	if strings.Contains(text, "\x1b[2K") {
		t.Errorf("a cursor code passed: %q", text)
	}
}

func TestRunStaysPlainWhenTheHostSaysPlain(t *testing.T) {
	handOn(t, false)
	t.Setenv("FORCE_COLOR", "1") // inherited from the person's shell
	log := captureLog(t)
	if err := Run(t.TempDir(), helperTool(t)); err != nil {
		t.Fatal(err)
	}
	if text := log.String(); strings.Contains(text, "\x1b") || !strings.Contains(text, "FORCE_COLOR=0 NO_COLOR=1") {
		t.Errorf("plain run: %q", text)
	}
}

func TestWhatAPluginReadsIsAlwaysAskedForPlainText(t *testing.T) {
	handOn(t, true)
	t.Setenv("FORCE_COLOR", "1")
	out, err := Output(t.TempDir(), helperTool(t))
	if err != nil || !strings.Contains(out, "FORCE_COLOR=0 NO_COLOR=1") {
		t.Errorf("Output: %v %q", err, out)
	}
	out, _, err = Capture(t.TempDir(), helperTool(t))
	if err != nil || !strings.Contains(out, "FORCE_COLOR=0 NO_COLOR=1") {
		t.Errorf("Capture: %v %q", err, out)
	}
	out, _, err = CaptureTimeout(t.TempDir(), helperTool(t), 30e9)
	if err != nil || !strings.Contains(out, "FORCE_COLOR=0 NO_COLOR=1") {
		t.Errorf("CaptureTimeout: %v %q", err, out)
	}
}
