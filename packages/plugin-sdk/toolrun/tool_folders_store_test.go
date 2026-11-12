package toolrun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordAndReadToolFolders(t *testing.T) {
	file := ToolFoldersFile(t.TempDir())
	got, err := ReadToolFolders(file)
	if err != nil || len(got) != 0 {
		t.Fatalf("a missing file reads as empty: %v %v", got, err)
	}
	if err := RecordToolFolder(file, "dart", `C:\pkg\dart-sdk\bin`); err != nil {
		t.Fatal(err)
	}
	if err := RecordToolFolder(file, "php", `C:\pkg\php`); err != nil {
		t.Fatal(err)
	}
	got, err = ReadToolFolders(file)
	if err != nil || got["dart"] != `C:\pkg\dart-sdk\bin` || got["php"] != `C:\pkg\php` {
		t.Fatalf("recorded folders: %v %v", got, err)
	}
}

func TestRecordToolFolderNeedsAHome(t *testing.T) {
	if ToolFoldersFile("") != "" {
		t.Fatal("no home, no file")
	}
	if err := RecordToolFolder("", "dart", "x"); err == nil {
		t.Fatal("recording without a home must fail")
	}
}

func TestWithToolFoldersAppendsOnlyNewFolders(t *testing.T) {
	sep := string(os.PathListSeparator)
	a, b, c := filepath.Join("x", "a"), filepath.Join("x", "b"), filepath.Join("x", "c")
	got := WithToolFolders(a+sep+b, map[string]string{"dart": c, "php": b, "none": ""})
	if got != a+sep+b+sep+c {
		t.Fatalf("PATH = %q", got)
	}
	if got := WithToolFolders("", map[string]string{"dart": c}); !strings.HasSuffix(got, c) {
		t.Fatalf("empty PATH = %q", got)
	}
}

func TestUseToolFoldersSkipsAFolderThatIsGone(t *testing.T) {
	home := t.TempDir()
	file := ToolFoldersFile(home)
	there := filepath.Join(home, "there")
	if err := os.Mkdir(there, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = RecordToolFolder(file, "dart", there)
	_ = RecordToolFolder(file, "php", filepath.Join(home, "gone"))
	t.Setenv("PATH", filepath.Join(home, "first"))
	if err := UseToolFolders(file); err != nil {
		t.Fatal(err)
	}
	path := os.Getenv("PATH")
	if !strings.HasSuffix(path, there) || strings.Contains(path, "gone") {
		t.Fatalf("PATH = %q", path)
	}
}
