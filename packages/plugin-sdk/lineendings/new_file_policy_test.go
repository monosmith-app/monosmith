package lineendings

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// repo makes a git repository with the given local config, ignoring the
// machine's own git configuration so the result is the same everywhere.
func repo(t *testing.T, config ...string) string {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Skipf("git is not available: %v %s", err, out)
	}
	for i := 0; i+1 < len(config); i += 2 {
		if out, err := exec.Command("git", "-C", dir, "config", config[i], config[i+1]).CombinedOutput(); err != nil {
			t.Fatalf("git config: %v %s", err, out)
		}
	}
	return dir
}

func put(t *testing.T, dir, name, text string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNewFileCRLFAttributesWin(t *testing.T) {
	dir := repo(t, "core.autocrlf", "true")
	put(t, dir, ".gitattributes", "* text=auto eol=lf\n")
	if NewFileCRLF(dir, "package.json") {
		t.Error("an eol=lf rule must beat core.autocrlf=true")
	}
	put(t, dir, ".gitattributes", "*.yml eol=crlf\n")
	if !NewFileCRLF(dir, ".github/workflows/ci.yml") {
		t.Error("an eol=crlf rule must give CRLF")
	}
}

func TestNewFileCRLFGitSettings(t *testing.T) {
	if !NewFileCRLF(repo(t, "core.autocrlf", "true"), "a.json") {
		t.Error("core.autocrlf=true must give CRLF")
	}
	if NewFileCRLF(repo(t, "core.autocrlf", "input"), "a.json") {
		t.Error("core.autocrlf=input must give LF")
	}
	if !NewFileCRLF(repo(t, "core.eol", "crlf", "core.autocrlf", "input"), "a.json") {
		t.Error("core.eol=crlf must beat core.autocrlf")
	}
}

func TestNewFileCRLFIgnoresGitSettingsOutsideARepository(t *testing.T) {
	global := filepath.Join(t.TempDir(), "gitconfig")
	put(t, filepath.Dir(global), "gitconfig", "[core]\n\tautocrlf = true\n")
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(global))
	if NewFileCRLF(t.TempDir(), "a.json") {
		t.Error("a folder outside any repository is not converted by checkout")
	}
}

func TestNewFileCRLFFollowsSiblings(t *testing.T) {
	dir := repo(t)
	put(t, dir, "smith.toml", "a = 1\r\nb = 2\r\n")
	put(t, dir, "README.md", "x\r\ny\r\n")
	if !NewFileCRLF(dir, "package.json") {
		t.Error("CRLF files in the root must give CRLF")
	}
	put(t, dir, "apps/web/a.ts", "x\ny\n")
	if NewFileCRLF(dir, "apps/web/b.ts") {
		t.Error("an LF folder must give LF, whatever the root uses")
	}
	if NewFileCRLF(repo(t), "package.json") {
		t.Error("no signal at all must give LF")
	}
}
