package packagemanifests

import (
	"strings"
	"testing"
)

func TestRenderAUR(t *testing.T) {
	r := release()
	r.Version = "1.2.3-rc.1"
	r.Description = "It's a tool"
	r.Archives = append(r.Archives, Archive{Name: "my-tool_1.2.3-rc.1_linux_arm64.zip", OS: "linux", Arch: "arm64", SHA256: strings.Repeat("cd", 32)})
	files, skipped := Render(r, []string{"aur"})
	if len(skipped) != 0 || len(files) != 2 {
		t.Fatalf("%v %v", keys(files), skipped)
	}
	sum := strings.Repeat("ab", 32)
	pkgbuild := files["my-tool-bin.PKGBUILD"]
	for _, want := range []string{
		"pkgname=my-tool-bin", "pkgver=1.2.3_rc.1", `pkgdesc='It'\''s a tool'`, "arch=('x86_64' 'aarch64')",
		"url='https://github.com/Acme/tools'", "license=('MIT')", "provides=('my-tool')", "conflicts=('my-tool')",
		"source_x86_64=('https://github.com/Acme/tools/releases/download/my-tool%401.2.3/my-tool_1.2.3_linux_amd64.zip')",
		"sha256sums_x86_64=('" + sum + "')", "sha256sums_aarch64=('" + strings.Repeat("cd", 32) + "')",
		`install -Dm755 "$srcdir/mytool" "$pkgdir/usr/bin/mytool"`,
	} {
		if !strings.Contains(pkgbuild, want) {
			t.Errorf("PKGBUILD lacks %q:\n%s", want, pkgbuild)
		}
	}
	srcinfo := files["my-tool-bin.SRCINFO"]
	for _, want := range []string{"pkgbase = my-tool-bin\n", "\tpkgver = 1.2.3_rc.1\n", "\tarch = aarch64\n", "\tsha256sums_x86_64 = " + sum + "\n", "\npkgname = my-tool-bin\n"} {
		if !strings.Contains(srcinfo, want) {
			t.Errorf(".SRCINFO lacks %q:\n%s", want, srcinfo)
		}
	}
}

func TestRenderAURSkipsWithoutLinuxArchive(t *testing.T) {
	r := release()
	r.Archives = r.Archives[:2] // macOS only
	files, skipped := Render(r, []string{"aur"})
	if len(files) != 0 || len(skipped) != 1 || !strings.Contains(skipped[0], "no Linux archive") {
		t.Fatalf("%v %v", keys(files), skipped)
	}
}
