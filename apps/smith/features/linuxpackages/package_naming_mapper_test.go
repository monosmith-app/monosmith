package linuxpackages

import "testing"

func TestFileNameFollowsEachFormat(t *testing.T) {
	cases := []struct{ kind, version, arch, want string }{
		{"deb", "1.0.0", "amd64", "monosmith_1.0.0_amd64.deb"},
		{"deb", "1.1.0-rc.1", "arm64", "monosmith_1.1.0~rc.1_arm64.deb"},
		{"rpm", "1.0.0", "amd64", "monosmith-1.0.0-1.x86_64.rpm"},
		{"rpm", "2.0.0-beta.2+abc", "arm64", "monosmith-2.0.0~beta.2-1.aarch64.rpm"},
	}
	for _, c := range cases {
		got, err := FileName(c.kind, Package{Name: "monosmith", Version: c.version, Arch: c.arch})
		if err != nil || got != c.want {
			t.Errorf("FileName(%s, %s, %s) = %q, %v; want %q", c.kind, c.version, c.arch, got, err, c.want)
		}
	}
	if _, err := FileName("apk", Package{Name: "x", Version: "1.0.0", Arch: "amd64"}); err == nil {
		t.Error("an unknown kind is accepted")
	}
}
