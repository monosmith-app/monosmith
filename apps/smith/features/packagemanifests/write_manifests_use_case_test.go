package packagemanifests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteDescribesTheBuiltArchives(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "dist", "release", "cli")
	_ = os.MkdirAll(dir, 0o755)
	sum := strings.Repeat("0f", 32)
	_ = os.WriteFile(filepath.Join(dir, "checksums.txt"), []byte(sum+"  cli_1.0.0_darwin_arm64.zip\n"+sum+"  cli_1.0.0_windows_amd64.zip\n"), 0o644)
	built := Built{
		Assets:    []string{"dist/release/cli/cli_1.0.0_darwin_arm64.zip", "dist/release/cli/cli_1.0.0_windows_amd64.zip", "dist/release/cli/checksums.txt"},
		Platforms: map[string]string{"dist/release/cli/cli_1.0.0_darwin_arm64.zip": "darwin/arm64", "dist/release/cli/cli_1.0.0_windows_amd64.zip": "windows/amd64"},
	}
	r := Release{Product: "cli", Binary: "cli", Version: "1.0.0", Owner: "acme", Repo: "tools", Tag: "cli@1.0.0"}
	written, skipped, err := Write(root, built, r, []string{"homebrew", "scoop"})
	if err != nil || len(skipped) != 0 {
		t.Fatal(err, skipped)
	}
	if strings.Join(written, ",") != "dist/release/cli/cli.json,dist/release/cli/cli.rb" {
		t.Fatalf("written %v", written)
	}
	formula, _ := os.ReadFile(filepath.Join(dir, "cli.rb"))
	if !strings.Contains(string(formula), `sha256 "`+sum+`"`) || !strings.Contains(string(formula), "cli%401.0.0/cli_1.0.0_darwin_arm64.zip") {
		t.Fatalf("formula:\n%s", formula)
	}
	if _, skipped, err := Write(root, Built{Assets: []string{"dist/x.zip"}}, r, []string{"scoop"}); err == nil || len(skipped) != 1 {
		t.Fatal("no checksums, no manifests")
	}
}
