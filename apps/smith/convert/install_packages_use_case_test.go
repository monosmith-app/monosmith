package convert

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func nxFolder(t *testing.T, installed bool) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "nx.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if installed {
		if err := os.MkdirAll(filepath.Join(dir, "node_modules", "nx"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestNxNeedsInstallOnlyWhenItsPackagesAreMissing(t *testing.T) {
	if !NxNeedsInstall(nxFolder(t, false)) {
		t.Error("an Nx folder without node_modules needs the install")
	}
	if NxNeedsInstall(nxFolder(t, true)) {
		t.Error("an installed Nx folder does not")
	}
	if NxNeedsInstall(t.TempDir()) {
		t.Error("a folder without nx.json does not")
	}
}

func TestEnsureInstalledAsksInATerminalAndInstallsOnYes(t *testing.T) {
	for _, answer := range []string{"\n", "y\n", "Yes\n"} {
		var out bytes.Buffer
		ran := 0
		err := EnsureInstalled(context.Background(), nxFolder(t, false), InstallChoice{
			Terminal: true, In: strings.NewReader(answer), Out: &out,
			Install: func(context.Context, string) error { ran++; return nil },
		})
		if err != nil || ran != 1 {
			t.Fatalf("answer %q: err %v, installs %d", answer, err, ran)
		}
		for _, want := range []string{"couple of minutes", "[Y/n]", "Installing the packages", "Packages installed in"} {
			if !strings.Contains(out.String(), want) {
				t.Errorf("the screen lacks %q:\n%s", want, out.String())
			}
		}
	}
}

func TestEnsureInstalledStopsOnNoAndWhenNobodyCanAnswer(t *testing.T) {
	ran := 0
	install := func(context.Context, string) error { ran++; return nil }
	err := EnsureInstalled(context.Background(), nxFolder(t, false), InstallChoice{Terminal: true, In: strings.NewReader("n\n"), Out: &bytes.Buffer{}, Install: install})
	if err == nil || ran != 0 {
		t.Fatalf("no: err %v, installs %d", err, ran)
	}
	err = EnsureInstalled(context.Background(), nxFolder(t, false), InstallChoice{Out: &bytes.Buffer{}, Install: install})
	if err == nil || ran != 0 || !strings.Contains(err.Error(), "must be installed first") {
		t.Fatalf("no terminal: err %v, installs %d", err, ran)
	}
}

func TestEnsureInstalledUnderYesInstallsWithoutAskingAndSkipsAReadyWorkspace(t *testing.T) {
	ran := 0
	install := func(context.Context, string) error { ran++; return nil }
	if err := EnsureInstalled(context.Background(), nxFolder(t, false), InstallChoice{Yes: true, Out: &bytes.Buffer{}, Install: install}); err != nil || ran != 1 {
		t.Fatalf("--yes: err %v, installs %d", err, ran)
	}
	if err := EnsureInstalled(context.Background(), nxFolder(t, true), InstallChoice{Yes: true, Out: &bytes.Buffer{}, Install: install}); err != nil || ran != 1 {
		t.Fatalf("ready: err %v, installs %d", err, ran)
	}
}
