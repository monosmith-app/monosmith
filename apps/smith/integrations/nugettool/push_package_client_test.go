package nugettool

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPushSendsThePackageWithTheAPIKey(t *testing.T) {
	file := filepath.Join(t.TempDir(), "acme.5.0.0.nupkg")
	_ = os.WriteFile(file, []byte("nupkg bytes"), 0o644)
	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, header, err := r.FormFile("package")
		if err != nil {
			t.Error(err)
			return
		}
		raw, _ := io.ReadAll(f)
		got = r.Method + " " + r.Header.Get("X-NuGet-ApiKey") + " " + r.Header.Get("X-NuGet-Client-Version") + " " + header.Filename + " " + string(raw)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	p := Pusher{URL: server.URL, APIKey: "key", Client: server.Client()}
	if err := p.Push(Package{ID: "acme", File: file}); err != nil {
		t.Fatal(err)
	}
	if got != "PUT key "+clientVersion+" acme.5.0.0.nupkg nupkg bytes" {
		t.Errorf("request = %q", got)
	}
}

func TestPushReportsAVersionTheFeedHasAndOtherFailures(t *testing.T) {
	file := filepath.Join(t.TempDir(), "acme.5.0.0.nupkg")
	_ = os.WriteFile(file, []byte("x"), 0o644)
	status := http.StatusConflict
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte("the API key cannot push acme"))
	}))
	defer server.Close()
	p := Pusher{URL: server.URL, APIKey: "key", Client: server.Client()}
	if err := p.Push(Package{File: file}); !errors.Is(err, ErrExists) {
		t.Errorf("409: err = %v", err)
	}
	status = http.StatusForbidden
	if err := p.Push(Package{File: file}); err == nil || !strings.Contains(err.Error(), "403 Forbidden: the API key cannot push acme") {
		t.Errorf("403: err = %v", err)
	}
}

func TestPusherFromEnvironmentNeedsTheAPIKey(t *testing.T) {
	t.Setenv("NUGET_API_KEY", "")
	t.Setenv("NUGET_PUSH_URL", "")
	if _, err := PusherFromEnvironment(); err == nil {
		t.Error("no error without NUGET_API_KEY")
	}
	t.Setenv("NUGET_API_KEY", "key")
	p, err := PusherFromEnvironment()
	if err != nil || p.URL != DefaultPushURL || p.APIKey != "key" {
		t.Errorf("pusher = %+v, %v", p, err)
	}
}
