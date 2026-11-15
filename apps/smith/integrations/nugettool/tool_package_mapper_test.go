package nugettool

import (
	"strings"
	"testing"
)

func TestRIDCoversEveryReleasePlatform(t *testing.T) {
	for platform, want := range map[string]string{"linux/amd64": "linux-x64", "darwin/arm64": "osx-arm64", "windows/arm64": "win-arm64"} {
		goos, goarch, _ := strings.Cut(platform, "/")
		if got, ok := RID(goos, goarch); !ok || got != want {
			t.Errorf("RID(%s) = %s, %v", platform, got, ok)
		}
	}
	if _, ok := RID("plan9", "amd64"); ok {
		t.Error("plan9 has a RID")
	}
}

func TestFileNameIsLowercase(t *testing.T) {
	if got := FileName("MonoSmith.linux-x64", "1.0.0-rc.1"); got != "monosmith.linux-x64.1.0.0-rc.1.nupkg" {
		t.Errorf("FileName = %s", got)
	}
}

func TestNuspecEscapesText(t *testing.T) {
	got := nuspec(Metadata{ID: "acme", Version: "1.0.0", Authors: "A & B", Description: "<tasks>"}, "DotnetTool")
	if !strings.Contains(got, "<authors>A &amp; B</authors>") || !strings.Contains(got, "<description>&lt;tasks&gt;</description>") {
		t.Errorf("nuspec = %s", got)
	}
}

func TestContentTypesListEachExtensionOnceAndOverrideFilesWithoutOne(t *testing.T) {
	got := contentTypes([]string{"acme.nuspec", "tools/any/osx-x64/acme", "tools/any/osx-x64/DotnetToolSettings.xml", "README.md"})
	for _, want := range []string{`<Default Extension="nuspec"`, `<Default Extension="xml"`, `<Default Extension="md"`, `<Override PartName="/tools/any/osx-x64/acme"`} {
		if strings.Count(got, want) != 1 {
			t.Errorf("content types have %s %d times:\n%s", want, strings.Count(got, want), got)
		}
	}
}
