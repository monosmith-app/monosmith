package releasing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestWriteVersionChangesOnlyTheTopLevelVersion(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "apps", "x")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	in := "{\n  \"engine\": {\"version\": \"9\"},\n  \"version\": \"1.0.0\",\n  \"name\": \"x\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "app.json"), []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := WriteVersion(ws, protocol.VersionParams{Project: protocol.Project{Root: "apps/x"}, Version: "1.1.0"})
	if err != nil || len(res.Files) != 1 {
		t.Fatalf("got %+v, %v", res, err)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "app.json"))
	want := "{\n  \"engine\": {\"version\": \"9\"},\n  \"version\": \"1.1.0\",\n  \"name\": \"x\"\n}\n"
	if string(got) != want {
		t.Fatalf("got %q", got)
	}
}
