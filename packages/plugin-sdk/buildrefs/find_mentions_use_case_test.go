package buildrefs

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMentionsFindsTheFilesThatNameAnOldBuildFolder(t *testing.T) {
	root := t.TempDir()
	write(t, root, "smith.toml", "command = [\"cp\", \"src/App/bin/Release/app.dll\", \"x\"]\n")
	write(t, root, ".github/workflows/ci.yml", "path: src/App/obj/\n")
	write(t, root, "src/App/Dockerfile", "COPY bin/Release/net8.0/publish/ .\n")
	write(t, root, "src/App/notes.sh", "echo robin/Release\n")
	write(t, root, "src/Other/Dockerfile", "COPY bin/Release .\n")
	write(t, root, "Makefile", "build:\n\tdotnet build src/AppX/bin\n")
	got := Mentions(root, "src/App", "bin", "obj")
	want := []string{".github/workflows/ci.yml", "smith.toml", "src/App/Dockerfile"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestMentionsIsEmptyWhenNothingNamesTheFolder(t *testing.T) {
	root := t.TempDir()
	write(t, root, "smith.toml", "[projects.app]\n")
	if got := Mentions(root, ".", "target"); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}
