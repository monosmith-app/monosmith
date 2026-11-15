package pubpackage

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func archive(t *testing.T, root, name, file string) string {
	t.Helper()
	f, err := os.Create(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create(file)
	_, _ = w.Write([]byte("binary " + file))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return name
}

func release(t *testing.T, root string) Release {
	_ = os.WriteFile(filepath.Join(root, "LICENSE"), []byte("MIT"), 0o644)
	return Release{
		Name: "acme", Binary: "ac", Version: "1.0.0", Description: "Builds things", License: "LICENSE", OutDir: "pub",
		Archives: []Archive{
			{Path: archive(t, root, "ac_windows_amd64.zip", "ac.exe"), OS: "windows", Arch: "amd64"},
			{Path: archive(t, root, "ac_darwin_arm64.zip", "ac"), OS: "darwin", Arch: "arm64"},
			{Path: archive(t, root, "ac_freebsd_amd64.zip", "ac"), OS: "freebsd", Arch: "amd64"},
		},
	}
}

// fake replaces pub.dev and the Dart CLI for one test.
func fake(t *testing.T, versions []string, run func(dir string, args ...string) (string, string, error)) {
	t.Helper()
	originalDart, originalPublished := dart, published
	dart, published = run, func(string) ([]string, error) { return versions, nil }
	t.Cleanup(func() { dart, published = originalDart, originalPublished })
}

func TestPublishBuildsThePackageAndPublishesIt(t *testing.T) {
	root := t.TempDir()
	var ran []string
	fake(t, []string{"0.9.0"}, func(dir string, args ...string) (string, string, error) {
		ran = append(ran, strings.Join(args, " "))
		// It publishes a copy outside the workspace, whose dist/ git ignores.
		if strings.HasPrefix(dir, root) {
			t.Errorf("published from %s, inside the workspace", dir)
		}
		if _, err := os.Stat(filepath.Join(dir, "pubspec.yaml")); err != nil {
			t.Errorf("the published copy has no pubspec.yaml")
		}
		return "", "", nil
	})
	res, err := Publish(root, release(t, root), false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Published != "acme@1.0.0" || len(res.Skipped) != 1 || strings.Join(ran, ";") != "pub publish --force" {
		t.Errorf("result %+v, ran %v", res, ran)
	}
	for _, file := range []string{"pubspec.yaml", "LICENSE", "CHANGELOG.md", "README.md", "bin/ac.dart", "lib/acme.dart", "lib/src/bin/windows-x64/ac.exe", "lib/src/bin/macos-arm64/ac"} {
		if _, err := os.Stat(filepath.Join(root, "pub", filepath.FromSlash(file))); err != nil {
			t.Errorf("the package has no %s", file)
		}
	}
}

func TestPublishShowsDartsValidationErrors(t *testing.T) {
	root := t.TempDir()
	fake(t, nil, func(string, ...string) (string, string, error) {
		return "* You must have a LICENSE file in the root directory.", "", errors.New("exit status 65")
	})
	_, err := Publish(root, release(t, root), false)
	if err == nil || !strings.Contains(err.Error(), "You must have a LICENSE file") {
		t.Errorf("err = %v", err)
	}
}

func TestPublishLeavesAVersionPubDevHas(t *testing.T) {
	root := t.TempDir()
	fake(t, []string{"1.0.0"}, func(string, ...string) (string, string, error) {
		t.Error("published a version pub.dev has")
		return "", "", nil
	})
	res, err := Publish(root, release(t, root), false)
	if err != nil || res.Existing != "acme@1.0.0" {
		t.Errorf("result %+v, %v", res, err)
	}
}
