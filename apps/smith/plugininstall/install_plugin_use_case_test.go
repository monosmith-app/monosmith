package plugininstall

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginarchive"
)

func entry(name, body string, exec bool) pluginarchive.Entry {
	return pluginarchive.Entry{Name: name, Executable: exec, Size: int64(len(body)), Open: func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(body)), nil
	}}
}

// archive packs a one-program plugin and returns its bytes and digest.
func archive(t *testing.T, extra ...pluginarchive.Entry) ([]byte, string) {
	t.Helper()
	entries := append([]pluginarchive.Entry{
		entry("plugin.json", `{"name":"widget","version":"0.1.0"}`, false),
		entry("bin/widget", "program", true),
	}, extra...)
	var buf bytes.Buffer
	if _, err := pluginarchive.Write(&buf, entries); err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(buf.Bytes())
	return buf.Bytes(), hex.EncodeToString(h[:])
}

func inspectWidget(dir string) (Plugin, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "plugin.json"))
	if err != nil {
		return Plugin{}, err
	}
	var m Plugin
	var shape struct{ Name, Version string }
	if err := json.Unmarshal(raw, &shape); err != nil {
		return Plugin{}, err
	}
	m.Name, m.Version = shape.Name, shape.Version
	return m, nil
}

func writeFile(t *testing.T, dir, name string, body []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, body, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func request(source, pluginsDir string) Request {
	return Request{Source: source, PluginsDir: pluginsDir, Inspect: inspectWidget}
}

func TestInstallFromALocalArchiveVerifiesTheDigest(t *testing.T) {
	raw, digest := archive(t)
	path := writeFile(t, t.TempDir(), "widget-0.1.0.smith-plugin.zip", raw)
	plugins := filepath.Join(t.TempDir(), "plugins")

	req := request(path, plugins)
	req.SHA256 = "sha256:" + strings.ToUpper(digest)
	res, err := Install(context.Background(), NewClient(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Name != "widget" || res.Dir != filepath.Join(plugins, "widget") || !res.ArchiveVerified || !res.Checksummed || res.Files != 2 {
		t.Fatalf("got %+v", res)
	}
	if got, _ := os.ReadFile(filepath.Join(res.Dir, "bin", "widget")); string(got) != "program" {
		t.Fatalf("installed content %q", got)
	}
	if left, _ := filepath.Glob(filepath.Join(plugins, ".*")); len(left) != 0 {
		t.Fatalf("staging left behind: %v", left)
	}
}

func TestInstallRefusesADigestMismatchAndInstallsNothing(t *testing.T) {
	raw, _ := archive(t)
	path := writeFile(t, t.TempDir(), "w.smith-plugin.zip", raw)
	plugins := t.TempDir()
	req := request(path, plugins)
	req.SHA256 = strings.Repeat("0", 64)
	_, err := Install(context.Background(), NewClient(), req)
	var mismatch *DigestMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("got %v", err)
	}
	if entries, _ := os.ReadDir(plugins); len(entries) != 0 {
		t.Fatalf("something was installed or left behind: %v", entries)
	}
}

func TestInstallRefusesAnArchiveWithoutChecksums(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("plugin.json")
	_, _ = w.Write([]byte(`{"name":"widget","version":"1"}`))
	_ = zw.Close()
	path := writeFile(t, t.TempDir(), "w.smith-plugin.zip", buf.Bytes())
	_, err := Install(context.Background(), NewClient(), request(path, t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), pluginarchive.ChecksumsFile) {
		t.Fatalf("got %v", err)
	}
}

func TestInstallFromADirectoryNeedsNoChecksums(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "plugin.json", []byte(`{"name":"widget","version":"0.1.0"}`))
	res, err := Install(context.Background(), NewClient(), request(src, t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	if res.Checksummed || res.ArchiveSHA256 != "" {
		t.Fatalf("got %+v", res)
	}
	req := request(src, t.TempDir())
	req.SHA256 = strings.Repeat("a", 64)
	if _, err := Install(context.Background(), NewClient(), req); err == nil || !strings.Contains(err.Error(), "directory") {
		t.Fatalf("--sha256 on a directory: %v", err)
	}
}

func TestInstallReplacesOnlyWhenAsked(t *testing.T) {
	raw, _ := archive(t)
	path := writeFile(t, t.TempDir(), "w.smith-plugin.zip", raw)
	plugins := t.TempDir()
	if _, err := Install(context.Background(), NewClient(), request(path, plugins)); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(plugins, "widget", "local-edit")
	if err := os.WriteFile(marker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(context.Background(), NewClient(), request(path, plugins)); !errors.Is(err, ErrAlreadyInstalled) {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("a refused install touched the installed plugin")
	}
	req := request(path, plugins)
	req.Replace = true
	res, err := Install(context.Background(), NewClient(), req)
	if err != nil || !res.Replaced {
		t.Fatalf("got %+v, %v", res, err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the old files are still there")
	}
	if left, _ := filepath.Glob(filepath.Join(plugins, ".*")); len(left) != 0 {
		t.Fatalf("left behind: %v", left)
	}
}

func TestInstallRefusesWhatInspectRefusesAndABadName(t *testing.T) {
	raw, _ := archive(t)
	path := writeFile(t, t.TempDir(), "w.smith-plugin.zip", raw)
	plugins := t.TempDir()
	req := request(path, plugins)
	req.Inspect = func(string) (Plugin, error) { return Plugin{}, fmt.Errorf("no program for this platform") }
	if _, err := Install(context.Background(), NewClient(), req); err == nil || !strings.Contains(err.Error(), "no program") {
		t.Fatalf("got %v", err)
	}
	req.Inspect = func(string) (Plugin, error) { return Plugin{Name: "../../evil"}, nil }
	if _, err := Install(context.Background(), NewClient(), req); err == nil {
		t.Fatal("installed under a traversal name")
	}
	if entries, _ := os.ReadDir(plugins); len(entries) != 0 {
		t.Fatalf("left behind: %v", entries)
	}
}

func TestInstallFromAURLWithADigestAndALimit(t *testing.T) {
	raw, digest := archive(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/w.smith-plugin.zip":
			_, _ = w.Write(raw)
		case "/redirect":
			http.Redirect(w, r, "http://example.com/w.smith-plugin.zip", http.StatusFound)
		case "/huge":
			// No Content-Length: the limit has to be enforced while reading.
			w.Header().Set("Transfer-Encoding", "chunked")
			for i := 0; i < 64; i++ {
				_, _ = w.Write(bytes.Repeat([]byte("x"), 1024))
				w.(http.Flusher).Flush()
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	req := request(srv.URL+"/w.smith-plugin.zip", t.TempDir())
	req.SHA256 = digest
	res, err := Install(context.Background(), NewClient(), req)
	if err != nil || !res.ArchiveVerified || res.Source.Kind != KindURL {
		t.Fatalf("got %+v, %v", res, err)
	}

	req = request(srv.URL+"/w.smith-plugin.zip", t.TempDir())
	req.SHA256 = strings.Repeat("1", 64)
	if _, err := Install(context.Background(), NewClient(), req); err == nil || !strings.Contains(err.Error(), "sha256") {
		t.Fatalf("wrong digest over http: %v", err)
	}

	if _, err := Install(context.Background(), NewClient(), request(srv.URL+"/redirect", t.TempDir())); err == nil || !strings.Contains(err.Error(), "only https") {
		t.Fatalf("a redirect to plain http: %v", err)
	}
	if _, err := Install(context.Background(), NewClient(), request(srv.URL+"/missing", t.TempDir())); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("not found: %v", err)
	}

	var sink bytes.Buffer
	if _, _, err := download(context.Background(), NewClient(), srv.URL+"/huge", "", 10_000, &sink); err == nil || !strings.Contains(err.Error(), "download limit") {
		t.Fatalf("a body over the limit: %v", err)
	}
	if sink.Len() > 10_001 {
		t.Fatalf("read %d bytes past the limit", sink.Len())
	}
}

func TestInstallFromAGitHubReleaseUsesThePublishedDigest(t *testing.T) {
	raw, digest := archive(t)
	published := digest
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/widget/releases/tags/v0.1.0":
			assets := []map[string]string{
				{"name": "widget-0.1.0.smith-plugin.zip", "browser_download_url": srv.URL + "/dl/archive"},
				{"name": "widget-0.1.0.smith-plugin.zip.sha256", "browser_download_url": srv.URL + "/dl/digest"},
				{"name": "notes.txt", "browser_download_url": srv.URL + "/dl/notes"},
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"assets": assets})
		case "/dl/archive":
			_, _ = w.Write(raw)
		case "/dl/digest":
			fmt.Fprintf(w, "%s  widget-0.1.0.smith-plugin.zip\n", published)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	t.Setenv("GITHUB_API_URL", srv.URL)

	res, err := Install(context.Background(), NewClient(), request("github:acme/widget@v0.1.0", t.TempDir()))
	if err != nil || !res.ArchiveVerified || res.Source.Kind != KindGitHub {
		t.Fatalf("got %+v, %v", res, err)
	}

	published = strings.Repeat("f", 64)
	plugins := t.TempDir()
	_, err = Install(context.Background(), NewClient(), request("github:acme/widget@v0.1.0", plugins))
	var mismatch *DigestMismatchError
	if !errors.As(err, &mismatch) || !strings.Contains(mismatch.From, "release") {
		t.Fatalf("a release whose digest disagrees: %v", err)
	}
	if entries, _ := os.ReadDir(plugins); len(entries) != 0 {
		t.Fatalf("installed anyway: %v", entries)
	}

	if _, err := Install(context.Background(), NewClient(), request("github:acme/widget@v9", t.TempDir())); err == nil || !strings.Contains(err.Error(), "no release v9") {
		t.Fatalf("unknown tag: %v", err)
	}
}
