package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverReadsAppJSON(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/music-fixer/flow.json", `{}`)
	write(t, ws, "apps/music-fixer/app.json", `{"name":"Music Fixer","version":"1.2.0","dependsOn":["videoplugins-music-sync"]}`)
	write(t, ws, "apps/lone/flow.json", `{}`)
	write(t, ws, "apps/rusty/flow.json", `{}`)
	write(t, ws, "apps/rusty/app.json", `{"version":"1.0.0"}`)
	write(t, ws, "apps/rusty/Cargo.toml", "")
	res, err := Discover(ws, []string{"apps/music-fixer/flow.json", "apps/lone/flow.json", "apps/rusty/flow.json", "flow.json"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 1 {
		t.Fatalf("want 1 project (no app.json, a Cargo.toml and the root are skipped), got %+v", res.Projects)
	}
	p := res.Projects[0]
	if p.Name != "music-fixer" || p.Version != "1.2.0" || p.Kind != "app" || len(p.Imports) != 1 || p.Imports[0] != "videoplugins-music-sync" || len(p.Exports) != 1 || p.Exports[0] != "Music Fixer" {
		t.Fatalf("unexpected project %+v", p)
	}
}

func TestDiscoverRejectsBrokenAppJSON(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/x/flow.json", `{}`)
	write(t, ws, "apps/x/app.json", `{nope`)
	if _, err := Discover(ws, []string{"apps/x/flow.json"}); err == nil {
		t.Fatal("want an error naming apps/x/app.json")
	}
}
