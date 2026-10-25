package session

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestLocalCacheDirIsWorkspaceRelativeUnlessAbsolute(t *testing.T) {
	root := filepath.Join(t.TempDir(), "my workspace")
	cfg := func(dir string) workspace.Config {
		var c workspace.Config
		c.Cache.Dir = dir
		return c
	}
	absolute := `/var/cache/smith`
	if runtime.GOOS == "windows" {
		absolute = `D:\cache\smith`
	}
	for _, tc := range []struct {
		name, dir, env, want string
	}{
		{"the default", "", "", filepath.Join(root, ".smith", "cache")},
		{"a relative setting with forward slashes", "build/cache", "", filepath.Join(root, "build", "cache")},
		{"an absolute setting", absolute, "", absolute},
		{"the environment wins over the setting", ".smith/cache", absolute, absolute},
		{"a relative environment value is workspace-relative", "", "other/dir", filepath.Join(root, "other", "dir")},
	} {
		got := LocalCacheDir(root, cfg(tc.dir), env(map[string]string{"SMITH_CACHE_DIR": tc.env}))
		if got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestWindowsDriveAndUNCCachePathsAreAbsolute(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("drive letters and UNC shares are Windows paths")
	}
	root := `C:\work\acme`
	for _, dir := range []string{`D:\smith cache`, `\\build-server\share\cache`, `D:/forward/slashes`} {
		var c workspace.Config
		c.Cache.Dir = dir
		got := LocalCacheDir(root, c, env(nil))
		if !filepath.IsAbs(got) || got != dir {
			t.Errorf("%s became %s", dir, got)
		}
	}
}

func TestCacheLimitBytes(t *testing.T) {
	for mb, want := range map[int]int64{2048: 2048 << 20, 1: 1 << 20, 0: 0, -1: 0} {
		var c workspace.Config
		c.Cache.MaxSizeMB = mb
		if got := CacheLimitBytes(c); got != want {
			t.Errorf("%d MB -> %d, want %d", mb, got, want)
		}
	}
}

func TestPlanRemote(t *testing.T) {
	var c workspace.Config
	if got := PlanRemote(c, env(nil)); got.URL != "" || got.Active {
		t.Fatalf("no remote: %+v", got)
	}
	c.Cache.Remote = "s3://bucket/prefix"
	if got := PlanRemote(c, env(nil)); got.URL != "s3://bucket/prefix" || got.Active {
		t.Fatalf("no key, so it is off: %+v", got)
	}
	if got := PlanRemote(c, env(map[string]string{"SMITH_CACHE_KEY": "secret"})); !got.Active || got.ReadOnly {
		t.Fatalf("with a key: %+v", got)
	}
	if got := PlanRemote(c, env(map[string]string{"SMITH_CACHE_KEY": "$(SMITH_CACHE_KEY)"})); got.Active {
		t.Fatalf("an Azure macro nobody expanded is no key: %+v", got)
	}
	if got := PlanRemote(c, env(map[string]string{"SMITH_CACHE_KEY": "k", "GITHUB_EVENT_NAME": "pull_request"})); !got.ReadOnly {
		t.Fatalf("a pull request only reads: %+v", got)
	}
	if got := PlanRemote(c, env(map[string]string{"SMITH_CACHE_KEY": "k", "SMITH_CACHE_REMOTE": "https://other.example/x", "SMITH_CACHE_READ_ONLY": "1"})); got.URL != "https://other.example/x" || !got.ReadOnly {
		t.Fatalf("the environment overrides: %+v", got)
	}
	c.Cache.RemoteReadOnly = true
	if got := PlanRemote(c, env(map[string]string{"SMITH_CACHE_KEY": "k"})); !got.ReadOnly {
		t.Fatalf("remote_read_only: %+v", got)
	}
}
