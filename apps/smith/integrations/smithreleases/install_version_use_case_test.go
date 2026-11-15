package smithreleases

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testdata is a release signed by smith release's own code with a test key,
// the same files the npm launcher's tests read.
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func testKey(t *testing.T) string {
	return strings.Split(string(fixture(t, "test-key.pub")), "\n")[1]
}

func mirror(t *testing.T) *httptest.Server {
	t.Helper()
	files := map[string][]byte{"/latest.txt": []byte("1.2.3\n"), "/next.txt": []byte("1.3.0-beta.1\n")}
	for _, name := range []string{"monosmith_1.2.3_linux_amd64.zip", "checksums.txt", "checksums.txt.minisig"} {
		files["/smith%401.2.3/"+name] = fixture(t, name)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := files[r.URL.RawPath]
		if !ok {
			body, ok = files[r.URL.Path]
		}
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("SMITH_DOWNLOAD_URL", srv.URL)
	return srv
}

func TestInstallPutsACheckedVersionInTheVersionsFolder(t *testing.T) {
	mirror(t)
	home := t.TempDir()
	program, err := Install(context.Background(), home, "1.2.3", "linux", "amd64", testKey(t))
	if err != nil {
		t.Fatal(err)
	}
	if program != filepath.Join(home, "versions", "1.2.3", "smith") {
		t.Fatalf("program at %s", program)
	}
	if raw, _ := os.ReadFile(program); string(raw) != "#!/bin/sh\necho smith 1.2.3\n" {
		t.Fatalf("program holds %q", raw)
	}
}

func TestInstallRefusesAReleaseAnotherKeySigned(t *testing.T) {
	mirror(t)
	home := t.TempDir()
	_, err := Install(context.Background(), home, "1.2.3", "linux", "amd64", PublicKey)
	if err == nil || !strings.Contains(err.Error(), "signature of checksums.txt does not match") {
		t.Fatalf("want a signature refusal, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "versions", "1.2.3")); !os.IsNotExist(err) {
		t.Fatal("nothing is written for a refused release")
	}
}

func TestNewestReadsTheMirrorForEachChannel(t *testing.T) {
	mirror(t)
	for channel, want := range map[string]string{"latest": "1.2.3", "next": "1.3.0-beta.1"} {
		if got, err := Newest(context.Background(), channel); err != nil || got != want {
			t.Errorf("%s: got %q, %v", channel, got, err)
		}
	}
}

func TestCheckNamesWhatIsWrong(t *testing.T) {
	archive, sums, sig := fixture(t, "monosmith_1.2.3_linux_amd64.zip"), fixture(t, "checksums.txt"), fixture(t, "checksums.txt.minisig")
	if why := Check(archive, "monosmith_1.2.3_linux_amd64.zip", sums, sig, testKey(t)); why != "" {
		t.Fatalf("a signed, listed archive is trusted: %s", why)
	}
	if why := Check([]byte("other"), "monosmith_1.2.3_linux_amd64.zip", sums, sig, testKey(t)); !strings.Contains(why, "does not match its checksum") {
		t.Errorf("changed archive: %s", why)
	}
	if why := Check(archive, "monosmith_1.2.3_darwin_arm64.zip", sums, sig, testKey(t)); !strings.Contains(why, "does not list") {
		t.Errorf("unlisted archive: %s", why)
	}
}
