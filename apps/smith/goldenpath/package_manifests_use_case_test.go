//go:build e2e

package goldenpath

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestReleaseAttachesPackageManifests releases a Go app to a (fake) GitHub
// Release and checks that its Homebrew, Scoop, winget and AUR manifests are
// attached beside the archives, built from their checksums.
func TestReleaseAttachesPackageManifests(t *testing.T) {
	var mu sync.Mutex
	uploads := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/releases/tags/"):
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPost && r.URL.Path == "/repos/acme/tools/releases":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "html_url": "https://example/rel", "upload_url": "http://" + r.Host + "/upload{?name,label}"})
		case r.Method == http.MethodPost && r.URL.Path == "/upload":
			raw, _ := io.ReadAll(r.Body)
			uploads[r.URL.Query().Get("name")] = string(raw)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusTeapot)
		}
	}))
	defer srv.Close()

	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: append(identity(), "GITHUB_TOKEN=tok", "GITHUB_API_URL="+srv.URL, "GITHUB_REPOSITORY=acme/tools")}
	w.smith(true, "new", "acme", "--ci", "github")
	w.dir = filepath.Join(base, "acme")
	w.smith(true, "add", "app", "tool", "--lang", "go")
	cfg := w.read("smith.toml")
	w.write("smith.toml", strings.Replace(cfg, "[ci]", "[release]\ngithub_releases = true\n\n[ci]", 1)+
		"\n[projects.tool.release]\nplatforms = [\"darwin/arm64\", \"linux/amd64\", \"windows/amd64\"]\npackage_managers = [\"homebrew\", \"scoop\", \"winget\", \"aur\"]\ndescription = \"A tool\"\nlicense = \"MIT\"\n")
	remote := filepath.Join(base, "remote.git")
	run(t, base, "git", "init", "-q", "--bare", remote)
	w.git("remote", "add", "origin", remote)
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: the tool")
	w.git("push", "-q", "origin", "main")

	out := w.smith(true, "release", "--yes")
	mu.Lock()
	defer mu.Unlock()
	for _, want := range []string{"tool_0.1.0_windows_amd64.zip", "checksums.txt", "tool.rb", "tool.json", "acme.tool.yaml", "acme.tool.installer.yaml", "acme.tool.locale.en-US.yaml", "tool-bin.PKGBUILD", "tool-bin.SRCINFO"} {
		if _, ok := uploads[want]; !ok {
			t.Fatalf("%s was not attached (attached: %v)\n%s", want, names(uploads), out)
		}
	}
	sums := uploads["checksums.txt"]
	for _, line := range strings.Split(strings.TrimSpace(sums), "\n") {
		sum, name, _ := strings.Cut(line, "  ")
		if strings.HasSuffix(name, "darwin_arm64.zip") && !strings.Contains(uploads["tool.rb"], sum) {
			t.Fatalf("the formula's sha256 must be the archive's:\n%s", uploads["tool.rb"])
		}
	}
	if !strings.Contains(uploads["acme.tool.installer.yaml"], "RelativeFilePath: tool.exe") || !strings.Contains(uploads["tool.json"], `"bin": "tool.exe"`) {
		t.Fatalf("windows manifests:\n%s\n%s", uploads["acme.tool.installer.yaml"], uploads["tool.json"])
	}
	if _, err := os.Stat(filepath.Join(w.dir, "dist", "release", "tool", "tool.rb")); err != nil {
		t.Fatal("the manifests are written beside the archives")
	}
}

func names(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
