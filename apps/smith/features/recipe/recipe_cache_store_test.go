package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLockKeepsOneSortedEntryPerUrl(t *testing.T) {
	root := t.TempDir()
	l := Lock{}.with("https://b.example/r.toml", "bb").with("https://a.example/r.toml", "aa").with("https://b.example/r.toml", "cc")
	if len(l.Recipes) != 2 || l.Recipes[0].URL != "https://a.example/r.toml" || l.Recipes[1].SHA256 != "cc" {
		t.Fatalf("lock = %+v", l)
	}
	if err := writeLock(root, l); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, LockFile))
	if !strings.HasPrefix(string(raw), "# Written by smith.") {
		t.Errorf("the lockfile must say who writes it:\n%s", raw)
	}
	back, err := readLock(root)
	if err != nil || len(back.Recipes) != 2 || back.Recipes[1] != l.Recipes[1] {
		t.Fatalf("round trip = %+v, %v", back, err)
	}
	if pin, ok := back.find("https://a.example/r.toml"); !ok || pin.SHA256 != "aa" {
		t.Errorf("find = %+v %v", pin, ok)
	}
}

func TestReadLockOfAWorkspaceWithNoneIsEmpty(t *testing.T) {
	l, err := readLock(t.TempDir())
	if err != nil || len(l.Recipes) != 0 {
		t.Errorf("lock = %+v, %v", l, err)
	}
}

func TestReadLockRefusesAFileThatIsNotTOML(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, LockFile, "recipe = [[[")
	if _, err := readLock(root); err == nil || !strings.Contains(err.Error(), LockFile) {
		t.Errorf("err = %v", err)
	}
}

func TestCacheKeepsAUrlsContentUnderTheWorkspace(t *testing.T) {
	root := t.TempDir()
	if readCache(root, "https://a.example/r.toml") != nil {
		t.Fatal("nothing is cached yet")
	}
	if err := writeCache(root, "https://a.example/r.toml", []byte("name = \"r\"\n")); err != nil {
		t.Fatal(err)
	}
	if got := string(readCache(root, "https://a.example/r.toml")); got != "name = \"r\"\n" {
		t.Errorf("cache = %q", got)
	}
	if !strings.HasPrefix(cachePath(root, "https://a.example/r.toml"), filepath.Join(root, ".smith", "recipes")) {
		t.Errorf("cache path = %s", cachePath(root, "https://a.example/r.toml"))
	}
}
