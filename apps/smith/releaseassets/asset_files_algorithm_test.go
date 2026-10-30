package releaseassets

import (
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestAssetFilesExpandsGlobsSortedWithoutRepeats(t *testing.T) {
	fsys := fstest.MapFS{"out/b.zip": {}, "out/a.zip": {}, "out/a.zip.sha256": {}, "out/sub/c.zip": {}}
	got, err := assetFiles(fsys, "/w", []string{"out/*.zip", "out/*.zip.sha256", "out/a.zip"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join("/w", "out", "a.zip"), filepath.Join("/w", "out", "a.zip.sha256"), filepath.Join("/w", "out", "b.zip")}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAssetFilesNamesAGlobTheBuildLeftNothingFor(t *testing.T) {
	_, err := assetFiles(fstest.MapFS{"out/a.txt": {}}, "/w", []string{"out/*.zip"})
	if err == nil || !strings.Contains(err.Error(), `"out/*.zip"`) {
		t.Errorf("got %v", err)
	}
}
