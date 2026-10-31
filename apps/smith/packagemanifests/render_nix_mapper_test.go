package packagemanifests

import (
	"strings"
	"testing"
)

func TestRenderNix(t *testing.T) {
	files, skipped := Render(release(), []string{"nix"})
	if len(skipped) != 0 || len(files) != 1 {
		t.Fatalf("%v %v", keys(files), skipped)
	}
	flake := files["my-tool.flake.nix"]
	for _, want := range []string{
		`pname = "my-tool";`, `version = "1.2.3";`,
		`aarch64-darwin = { url = "https://github.com/Acme/tools/releases/download/my-tool%401.2.3/my-tool_1.2.3_darwin_arm64.zip"; sha256 = "` + strings.Repeat("ab", 32) + `"; };`,
		"x86_64-linux = {", "install -Dm755 mytool $out/bin/mytool", `mainProgram = "mytool";`, "/bin/mytool\";",
	} {
		if !strings.Contains(flake, want) {
			t.Errorf("flake lacks %q:\n%s", want, flake)
		}
	}
	if strings.Contains(flake, "windows") || strings.Contains(flake, "aarch64-linux = {") {
		t.Errorf("flake names a platform without an archive:\n%s", flake)
	}
}

func TestRenderNixSkipsWithoutUnixArchive(t *testing.T) {
	r := release()
	r.Archives = r.Archives[3:] // Windows only
	files, skipped := Render(r, []string{"nix"})
	if len(files) != 0 || len(skipped) != 1 || !strings.Contains(skipped[0], "no macOS or Linux archive") {
		t.Fatalf("%v %v", keys(files), skipped)
	}
}
