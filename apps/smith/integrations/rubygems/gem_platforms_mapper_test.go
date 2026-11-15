package rubygems

import "testing"

func TestPlatformsGiveMuslItsOwnGem(t *testing.T) {
	got, ok := Platforms("linux", "arm64")
	if !ok || len(got) != 2 || got[0] != "aarch64-linux" || got[1] != "aarch64-linux-musl" {
		t.Errorf("linux/arm64 = %v %v", got, ok)
	}
	if got, _ := Platforms("windows", "amd64"); got[0] != "x64-mingw-ucrt" {
		t.Errorf("windows/amd64 = %v", got)
	}
	if _, ok := Platforms("freebsd", "amd64"); ok {
		t.Error("freebsd has no gem platform")
	}
}

func TestVersion(t *testing.T) {
	for in, want := range map[string]string{"1.2.0": "1.2.0", "1.2.0-rc.1": "1.2.0.pre.rc.1", "1.2.0-beta.2+abc": "1.2.0.pre.beta.2"} {
		if got, err := Version(in); err != nil || got != want {
			t.Errorf("Version(%s) = %s, %v; want %s", in, got, err, want)
		}
	}
	if _, err := Version("1.2.0-rc_1"); err == nil {
		t.Error("an underscore has no gem spelling")
	}
}

func TestFileName(t *testing.T) {
	if got := FileName("monosmith", "1.0.0", "arm64-darwin"); got != "monosmith-1.0.0-arm64-darwin.gem" {
		t.Error(got)
	}
}
