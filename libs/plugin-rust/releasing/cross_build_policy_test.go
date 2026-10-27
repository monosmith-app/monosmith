package releasing

import "testing"

func TestCrossReach(t *testing.T) {
	cases := []struct {
		host, platform string
		want           Reach
	}{
		{"linux/amd64", "linux/amd64", Here},
		{"windows/amd64", "windows/arm64", Here},
		{"darwin/arm64", "darwin/amd64", Here},
		{"linux/amd64", "linux/arm64", WithLinker},
		{"linux/arm64", "linux/amd64", WithLinker},
		{"linux/amd64", "windows/amd64", OnItsOwnSystem},
		{"darwin/arm64", "linux/arm64", OnItsOwnSystem},
		{"windows/amd64", "darwin/arm64", OnItsOwnSystem},
	}
	for _, c := range cases {
		if got := CrossReach(c.host, c.platform); got != c.want {
			t.Errorf("%s -> %s: %v, want %v", c.host, c.platform, got, c.want)
		}
	}
}

func TestLinkerNamesTheProgramAndItsPackage(t *testing.T) {
	program, pkg := Linker("linux/arm64")
	if program != "aarch64-linux-gnu-gcc" || pkg != "gcc-aarch64-linux-gnu" {
		t.Errorf("%s %s", program, pkg)
	}
	if _, pkg := Linker("linux/amd64"); pkg != "gcc-x86-64-linux-gnu" {
		t.Errorf("%s", pkg)
	}
	if got := LinkerEnv("aarch64-unknown-linux-gnu"); got != "CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER" {
		t.Error(got)
	}
}

func TestTargetCoversTheDefaultPlatforms(t *testing.T) {
	for _, p := range []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64", "windows/arm64"} {
		if _, err := Target(p); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
	if _, err := Target("freebsd/amd64"); err == nil {
		t.Error("an unknown platform has no target")
	}
}
