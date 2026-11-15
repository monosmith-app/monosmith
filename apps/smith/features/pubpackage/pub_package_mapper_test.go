package pubpackage

import (
	"strings"
	"testing"
)

func TestPlatformSpellsDartAbis(t *testing.T) {
	cases := map[[2]string][2]string{
		{"darwin", "arm64"}:  {"macos-arm64", "macosArm64"},
		{"windows", "amd64"}: {"windows-x64", "windowsX64"},
		{"linux", "arm64"}:   {"linux-arm64", "linuxArm64"},
	}
	for in, want := range cases {
		if dir, abi, ok := Platform(in[0], in[1]); !ok || dir != want[0] || abi != want[1] {
			t.Errorf("Platform(%v) = %s, %s, %v", in, dir, abi, ok)
		}
	}
	if _, _, ok := Platform("freebsd", "amd64"); ok {
		t.Error("Dart has no freebsd Abi")
	}
}

func TestPubspecDeclaresTheExecutable(t *testing.T) {
	r := Release{Name: "acme", Binary: "ac", Version: "1.0.0", Description: `Builds "things"`, Repository: "https://github.com/acme/acme"}
	got := Pubspec(r)
	for _, want := range []string{"name: acme\n", `description: "Builds \"things\""`, "version: 1.0.0\n", "repository: https://github.com/acme/acme\n", "executables:\n  ac: ac\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("pubspec lacks %q:\n%s", want, got)
		}
	}
}

func TestLauncherMapsEachAbiToItsBinary(t *testing.T) {
	got := Launcher(Release{Name: "acme", Binary: "ac"}, map[string]string{"windowsX64": "windows-x64", "linuxArm64": "linux-arm64"})
	for _, want := range []string{"  Abi.linuxArm64: 'linux-arm64',\n  Abi.windowsX64: 'windows-x64',\n", "package:acme/acme.dart", "'ac.exe' : 'ac'", "ProcessStartMode.inheritStdio"} {
		if !strings.Contains(got, want) {
			t.Errorf("launcher lacks %q:\n%s", want, got)
		}
	}
}

func TestDescriptionFitsPubDevLimits(t *testing.T) {
	if _, err := Description("too short"); err == nil {
		t.Error("a short description is accepted")
	}
	exact := strings.Repeat("word ", 20)
	if got, err := Description(exact); err != nil || got != strings.TrimSpace(exact) {
		t.Errorf("Description(100 chars) = %q, %v", got, err)
	}
	long := strings.Repeat("monorepo tasks, ", 20)
	got, err := Description(long)
	if err != nil || len(got) > 180 || len(got) < 60 || !strings.HasSuffix(got, "...") || strings.Contains(got, ", ...") {
		t.Errorf("Description(long) = %q (%d), %v", got, len(got), err)
	}
}
