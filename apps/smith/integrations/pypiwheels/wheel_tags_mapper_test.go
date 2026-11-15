package pypiwheels

import (
	"slices"
	"testing"
)

func TestTagsCoverGlibcAndMuslWithOneLinuxWheel(t *testing.T) {
	tags, ok := Tags("linux", "arm64")
	want := []string{"py3-none-manylinux_2_17_aarch64", "py3-none-manylinux2014_aarch64", "py3-none-musllinux_1_1_aarch64"}
	if !ok || !slices.Equal(tags, want) {
		t.Errorf("Tags = %v, %v", tags, ok)
	}
	if _, ok := Tags("freebsd", "amd64"); ok {
		t.Error("freebsd has no wheel platform")
	}
}

func TestFileNameCompressesTheTags(t *testing.T) {
	tags, _ := Tags("linux", "amd64")
	want := "my_tool-1.0.0-py3-none-manylinux_2_17_x86_64.manylinux2014_x86_64.musllinux_1_1_x86_64.whl"
	if got := FileName("My-Tool", "1.0.0", tags); got != want {
		t.Errorf("FileName = %s", got)
	}
}

func TestVersionSpellsPrereleasesForPython(t *testing.T) {
	cases := map[string]string{"5.0.0": "5.0.0", "1.2.0-rc.1": "1.2.0rc1", "1.2.0-beta.2": "1.2.0b2", "1.2.0-alpha.3": "1.2.0a3"}
	for in, want := range cases {
		if got, err := Version(in); err != nil || got != want {
			t.Errorf("Version(%s) = %s, %v; want %s", in, got, err, want)
		}
	}
	if _, err := Version("1.2.0-next.1"); err == nil {
		t.Error("a next prerelease has no Python spelling")
	}
}
