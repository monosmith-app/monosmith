package releaseassets

import (
	"path/filepath"
	"strings"
	"testing"
)

const tauriPattern = "{product}-{version}-{platform}.{ext}"

func TestAssetNamesRenamesTauriBundlesByPlatform(t *testing.T) {
	msi := filepath.Join("b", "msi", "VideoCraw Studio_0.1.0_x64_en-US.msi")
	exe := filepath.Join("b", "nsis", "VideoCraw Studio_0.1.0_x64-setup.exe")
	got, err := assetNames([]string{msi, exe}, tauriPattern, "studio", "0.1.0", "windows-x64")
	if err != nil {
		t.Fatal(err)
	}
	if got[msi] != "studio-0.1.0-windows-x64.msi" || got[exe] != "studio-0.1.0-windows-x64.exe" {
		t.Errorf("got %v", got)
	}
}

func TestAssetNamesKeepsTheBundleFolderWhenNamesCollide(t *testing.T) {
	a, b := filepath.Join("b", "nsis", "x-setup.exe"), filepath.Join("b", "wix", "y.exe")
	got, err := assetNames([]string{a, b}, tauriPattern, "studio", "1.0.0", "windows-x64")
	if err != nil {
		t.Fatal(err)
	}
	if got[a] != "studio-1.0.0-windows-x64-nsis.exe" || got[b] != "studio-1.0.0-windows-x64-wix.exe" {
		t.Errorf("got %v", got)
	}
}

func TestAssetNamesRefusesNamesThatStayTheSame(t *testing.T) {
	_, err := assetNames([]string{filepath.Join("a", "k", "x.deb"), filepath.Join("b", "k", "y.deb")}, tauriPattern, "s", "1", "linux-x64")
	if err == nil || !strings.Contains(err.Error(), "both be uploaded") {
		t.Errorf("got %v", err)
	}
}

func TestAssetNamesWithoutAPatternKeepsTheFileNames(t *testing.T) {
	f := filepath.Join("out", "a.zip")
	got, err := assetNames([]string{f}, "", "p", "1", "linux-x64")
	if err != nil || got[f] != "a.zip" {
		t.Errorf("got %v %v", got, err)
	}
}

func TestExtensionOfCountsCompoundExtensions(t *testing.T) {
	for file, want := range map[string]string{"a.tar.gz": "tar.gz", "a.AppImage": "AppImage", "a": "", "a.app.tar.gz": "app.tar.gz"} {
		if got := extensionOf(file); got != want {
			t.Errorf("%s: %q, want %q", file, got, want)
		}
	}
}
