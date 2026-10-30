package workspace

import (
	"strings"
	"testing"
)

const releaseHead = "[workspace]\nname = \"a\"\n[release]\ngithub_releases = true\n[projects.app.release]\n"

func TestParseReadsAReleaseBuildCommand(t *testing.T) {
	cfg, err := Parse([]byte(releaseHead + "build = \"make dist\"\nassets = [\"out/*.zip\"]\nplatforms = [\"windows-x64\", \"linux-arm64\"]\n"))
	if err != nil {
		t.Fatal(err)
	}
	r := cfg.Projects["app"].Release
	if r.Build != "make dist" || len(r.Assets) != 1 || len(r.Platforms) != 2 {
		t.Fatalf("got %+v", r)
	}
}

func TestParseRejectsAReleaseBuildThatCannotWork(t *testing.T) {
	for input, want := range map[string]string{
		releaseHead + "assets = [\"out/*.zip\"]\n":                                                                             "has no build",
		releaseHead + "build = \"make\"\nplatforms = [\"linux-x64\"]\n":                                                        "no assets",
		releaseHead + "build = \"make\"\nassets = [\"a\"]\n":                                                                   "no platforms",
		releaseHead + "build = \"make\"\nassets = [\"a\"]\nplatforms = [\"linux-x86\"]\n":                                      "use one of windows-x64",
		releaseHead + "build = \"make\"\nassets = [\"\"]\nplatforms = [\"linux-x64\"]\n":                                       "not a glob",
		releaseHead + "build = \"make\"\nassets = [\"a\"]\nplatforms = [\"linux-x64\"]\nnpm = [\"x\"]\n":                       "sets npm next to build",
		"[workspace]\nname = \"a\"\n[projects.app.release]\nbuild = \"make\"\nassets = [\"a\"]\nplatforms = [\"linux-x64\"]\n": "github_releases = true",
		releaseHead + "build = \"make\"\nassets = [\"a\"]\nplatforms = [\"linux-x64\"]\n[projects.app]\ntags = [\"native\"]\n": "tagged native",
	} {
		_, err := parseAndResolve(input)
		if err == nil || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "→") {
			t.Errorf("input %q: want a problem with a remedy containing %q, got %v", input, want, err)
		}
	}
}
