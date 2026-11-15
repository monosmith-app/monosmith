package clicontract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTheCacheLineShowsSizeAgainstLimit(t *testing.T) {
	for _, tc := range []struct {
		in   CacheInfo
		want string
	}{
		{CacheInfo{Bytes: 312 << 20, LimitBytes: 2048 << 20}, "312 MB of 2048 MB (15%)"},
		{CacheInfo{Bytes: 0, LimitBytes: 2048 << 20}, "0.0 MB of 2048 MB (0%)"},
		{CacheInfo{Bytes: 5 << 20}, "5 MB, no limit"},
		{CacheInfo{Bytes: 300 << 10, LimitBytes: 10 << 20}, "0.3 MB of 10 MB (3%)"},
	} {
		if got := tc.in.Line(); got != tc.want {
			t.Errorf("%+v -> %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestInfoReportsTheCacheOfTheWorkspace(t *testing.T) {
	ws := t.TempDir()
	t.Setenv("SMITH_NO_UPDATE_CHECK", "1")
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	t.Setenv("SMITH_CACHE_DIR", "")
	if err := os.WriteFile(filepath.Join(ws, "smith.toml"), []byte("schema = 1\n[workspace]\nname = \"acme\"\n[cache]\nmax_size_mb = 64\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	info := WorkspaceReport("dev", ws)
	if info.Workspace == nil || info.Workspace.Cache == nil {
		t.Fatalf("info %+v", info.Workspace)
	}
	c := info.Workspace.Cache
	if c.LimitBytes != 64<<20 || c.Bytes != 0 || c.Dir != filepath.Join(ws, ".smith", "cache") {
		t.Fatalf("cache %+v", c)
	}
	if _, err := os.Stat(filepath.Join(ws, ".smith")); err == nil {
		t.Fatal("smith info must not create the cache")
	}
}
