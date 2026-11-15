package releaseassets

import (
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestAssetFilesExpandsGlobsSortedWithoutRepeats(t *testing.T) {
	fsys := fstest.MapFS{"out/b.zip": {}, "out/a.zip": {}, "out/a.zip.sha256": {}, "out/sub/c.zip": {}}
	got, skipped, err := assetFiles(fsys, "/w", []string{"out/*.zip", "out/*.zip.sha256", "out/a.zip"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join("/w", "out", "a.zip"), filepath.Join("/w", "out", "a.zip.sha256"), filepath.Join("/w", "out", "b.zip")}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %v, want %v", got, want)
	}
	if len(skipped) != 0 {
		t.Errorf("skipped %v, want none", skipped)
	}
}

func TestAssetFilesAllowsAGlobThatMatchesNothingWhenAnotherMatched(t *testing.T) {
	fsys := fstest.MapFS{"dist/drop/mvd_1.0.0.msi": {}}
	got, skipped, err := assetFiles(fsys, "/w", []string{"dist/drop/mvd_*.msi", "dist/drop/mvd_*.dmg"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(skipped) != 1 || skipped[0] != "dist/drop/mvd_*.dmg" {
		t.Errorf("got %v, skipped %v", got, skipped)
	}
}

func TestAssetFilesNamesAGlobTheBuildLeftNothingFor(t *testing.T) {
	_, _, err := assetFiles(fstest.MapFS{"out/a.txt": {}}, "/w", []string{"out/*.zip"})
	if err == nil || !strings.Contains(err.Error(), `"out/*.zip"`) {
		t.Errorf("got %v", err)
	}
}

func TestAssetFilesFailsWhenNoGlobMatchedNamingAllOfThem(t *testing.T) {
	_, _, err := assetFiles(fstest.MapFS{}, "/w", []string{"out/*.msi", "out/*.dmg"})
	if err == nil || !strings.Contains(err.Error(), `"out/*.msi", "out/*.dmg"`) {
		t.Errorf("got %v", err)
	}
}
