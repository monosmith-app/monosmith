package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// options point the outputs of a fixture run into a temporary folder.
func options(t *testing.T, words string, check bool) Options {
	t.Helper()
	dir := t.TempDir()
	return Options{
		Words:              words,
		VSCodeModule:       filepath.Join(dir, "vscode", "src", "words"),
		VSCodeNls:          filepath.Join(dir, "vscode"),
		JetBrainsResources: filepath.Join(dir, "jb", "resources"),
		JetBrainsKotlin:    filepath.Join(dir, "jb", "kotlin"),
		Surfaces:           []string{"vscode", "jetbrains"},
		Check:              check,
	}
}

func TestSyncWritesEveryFileThenCheckFindsNothingStale(t *testing.T) {
	o := options(t, "fixtures", false)
	stale, err := Sync(o)
	if err != nil || len(stale) != 6 {
		t.Fatalf("first run: %v, %v", stale, err)
	}
	for _, p := range []string{
		filepath.Join(o.VSCodeModule, "smith-words.config.ts"),
		filepath.Join(o.VSCodeNls, "package.nls.json"),
		filepath.Join(o.VSCodeNls, "package.nls.pt-br.json"),
		filepath.Join(o.JetBrainsResources, "messages.properties"),
		filepath.Join(o.JetBrainsResources, "messages_pt_BR.properties"),
		filepath.Join(o.JetBrainsKotlin, "SmithWordsConfig.kt"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("not written: %v", err)
		}
	}
	o.Check = true
	if stale, err := Sync(o); err != nil || len(stale) != 0 {
		t.Fatalf("check after write: %v, %v", stale, err)
	}
}

func TestCheckFailsOnAChangedOrMissingFileAndWritesNothing(t *testing.T) {
	o := options(t, "fixtures", false)
	if _, err := Sync(o); err != nil {
		t.Fatal(err)
	}
	changed := filepath.Join(o.VSCodeNls, "package.nls.json")
	missing := filepath.Join(o.JetBrainsResources, "messages.properties")
	if err := os.WriteFile(changed, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(missing); err != nil {
		t.Fatal(err)
	}
	o.Check = true
	stale, err := Sync(o)
	if err != nil || !slices.Equal(stale, []string{changed, missing}) {
		t.Fatalf("stale = %v, %v", stale, err)
	}
	if _, err := os.Stat(missing); err == nil {
		t.Fatal("check mode must not write")
	}
}

func TestCheckIgnoresLineEndings(t *testing.T) {
	o := options(t, "fixtures", false)
	if _, err := Sync(o); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(o.VSCodeNls, "package.nls.json")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	o.Check = true
	if stale, err := Sync(o); err != nil || len(stale) != 0 {
		t.Fatalf("stale = %v, %v", stale, err)
	}
}

func TestEmptyKeyListsWriteValidFilesAndCheckToleratesAbsentOnes(t *testing.T) {
	o := options(t, filepath.Join("fixtures", "empty"), true)
	if stale, err := Sync(o); err != nil || len(stale) != 0 {
		t.Fatalf("check with nothing written: %v, %v", stale, err)
	}
	o.Check = false
	if stale, err := Sync(o); err != nil || len(stale) != 4 {
		t.Fatalf("write: %v, %v", stale, err)
	}
	if data, _ := os.ReadFile(filepath.Join(o.VSCodeNls, "package.nls.json")); string(data) != "{}\n" {
		t.Errorf("nls = %q", data)
	}
	o.Check = true
	if stale, err := Sync(o); err != nil || len(stale) != 0 {
		t.Fatalf("check after write: %v, %v", stale, err)
	}
}

func TestSyncFailsOnAnUnknownSurface(t *testing.T) {
	o := options(t, "fixtures", true)
	o.Surfaces = []string{"emacs"}
	if _, err := Sync(o); err == nil {
		t.Fatal("a surface with no manifest must be an error")
	}
}

// The generated files of the repository stay up to date: `smith run test` is the
// CI check, as for the flag clients. A surface with no key needs no files yet.
func TestTheRepositoryFilesAreUpToDate(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	under := func(p string) string { return filepath.Join(root, filepath.FromSlash(p)) }
	stale, err := Sync(Options{
		Words:              under("libs/words"),
		VSCodeModule:       under("apps/vscode/src/words"),
		VSCodeNls:          under("apps/vscode"),
		JetBrainsResources: under("apps/jetbrains/src/main/resources"),
		JetBrainsKotlin:    under("apps/jetbrains/src/main/kotlin/dev/monosmith/jetbrains/words"),
		Surfaces:           []string{"vscode", "jetbrains"},
		Check:              true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(stale) > 0 {
		t.Fatalf("not up to date: %v\nrun `go run ./wordgen` in libs/words", stale)
	}
}
