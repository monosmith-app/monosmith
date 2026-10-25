package pluginadmit

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setup is a smith home, a workspace and a plugin file, with CI off.
func setup(t *testing.T) (ws string, plugin Plugin, file string) {
	t.Helper()
	for _, v := range []string{"CI", "GITHUB_ACTIONS", "TF_BUILD"} {
		t.Setenv(v, "")
	}
	t.Setenv("SMITH_HOME", t.TempDir())
	t.Cleanup(func() { SetPrompter(nil) })
	SetPrompter(nil)
	ws = t.TempDir()
	file = filepath.Join(t.TempDir(), "widget-plugin")
	if err := os.WriteFile(file, []byte("v1"), 0o755); err != nil {
		t.Fatal(err)
	}
	return ws, Plugin{Name: "widget", Command: []string{file}, Source: ".smith/plugins", Where: "/x/widget", ManifestEnv: []string{"AWS_*"}}, file
}

func answer(text string) (*Prompter, *bytes.Buffer) {
	var out bytes.Buffer
	return &Prompter{Out: &out, In: bufio.NewReader(strings.NewReader(text))}, &out
}

func TestAdmitRefusesWithoutAPrompterAndNamesTheRemedy(t *testing.T) {
	ws, p, _ := setup(t)
	err := Admit(ws, []Plugin{p})
	if !errors.Is(err, ErrUntrusted) {
		t.Fatalf("got %v", err)
	}
	for _, want := range []string{"✖ the external plugin \"widget\" is not trusted yet, so smith did not start it", "→ run `smith plugin trust widget` to review it and trust it"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("message lacks %q:\n%s", want, err)
		}
	}
}

func TestAdmitAsksOnceAndRemembers(t *testing.T) {
	ws, p, _ := setup(t)
	prompter, out := answer("y\n")
	SetPrompter(prompter)
	if err := Admit(ws, []Plugin{p}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"widget"`, ".smith/plugins", "sha256 ", "AWS_*", "not sandboxed", "[y/N]"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("prompt lacks %q:\n%s", want, out)
		}
	}
	SetPrompter(nil)
	if err := Admit(ws, []Plugin{p}); err != nil {
		t.Fatalf("not remembered: %v", err)
	}
}

func TestAdmitDecliningRefuses(t *testing.T) {
	ws, p, _ := setup(t)
	prompter, _ := answer("\n")
	SetPrompter(prompter)
	if err := Admit(ws, []Plugin{p}); !errors.Is(err, ErrUntrusted) {
		t.Fatalf("got %v", err)
	}
}

func TestAdmitAsksAgainWhenTheFileChanges(t *testing.T) {
	ws, p, file := setup(t)
	prompter, _ := answer("y\n")
	SetPrompter(prompter)
	if err := Admit(ws, []Plugin{p}); err != nil {
		t.Fatal(err)
	}
	SetPrompter(nil)
	if err := os.WriteFile(file, []byte("v2"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := Admit(ws, []Plugin{p})
	if !errors.Is(err, ErrUntrusted) || !strings.Contains(err.Error(), "changed since you trusted it") {
		t.Fatalf("got %v", err)
	}
}

func TestAdmitInCIWantsAPinAndNeverAsks(t *testing.T) {
	ws, p, _ := setup(t)
	t.Setenv("CI", "true")
	prompter, out := answer("y\n")
	SetPrompter(prompter)

	err := Admit(ws, []Plugin{p})
	if !errors.Is(err, ErrUntrusted) || !strings.Contains(err.Error(), "is not pinned") || !strings.Contains(err.Error(), "smith plugin trust widget --pin") {
		t.Fatalf("unpinned: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("asked in CI: %s", out)
	}

	sum := sha256.Sum256([]byte("v1"))
	p.Pin = hex.EncodeToString(sum[:])
	if err := Admit(ws, []Plugin{p}); err != nil {
		t.Fatalf("pinned: %v", err)
	}
	p.Pin = strings.Repeat("0", 64)
	if err := Admit(ws, []Plugin{p}); err == nil || !strings.Contains(err.Error(), "does not match the sha256 pinned") {
		t.Fatalf("wrong pin: %v", err)
	}
}

func TestAdmitNeedsNothingForNoExternalPlugins(t *testing.T) {
	if err := Admit(t.TempDir(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestAdmitSaysWhenTheProgramIsMissing(t *testing.T) {
	ws, p, _ := setup(t)
	p.Command = []string{filepath.Join(t.TempDir(), "gone")}
	err := Admit(ws, []Plugin{p})
	if !errors.Is(err, ErrUntrusted) || !strings.Contains(err.Error(), "cannot be started") {
		t.Fatalf("got %v", err)
	}
}
