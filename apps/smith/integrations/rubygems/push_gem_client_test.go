package rubygems

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestPushSendsTheGemWithTheKey(t *testing.T) {
	var got struct{ path, key, kind, body string }
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got.path, got.key, got.kind, got.body = r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("Content-Type"), string(b)
		_, _ = io.WriteString(w, "Successfully registered gem: acme (1.0.0)")
	}))
	defer srv.Close()
	file := filepath.Join(t.TempDir(), "acme-1.0.0-x86_64-linux.gem")
	_ = os.WriteFile(file, []byte("gem"), 0o644)
	if err := (Pusher{Host: srv.URL + "/", Key: "rubygems_key", Client: srv.Client()}).Push(Gem{File: file}); err != nil {
		t.Fatal(err)
	}
	if got.path != "/api/v1/gems" || got.key != "rubygems_key" || got.kind != "application/octet-stream" || got.body != "gem" {
		t.Errorf("request = %+v", got)
	}
}

func TestPushTreatsARepushAsExisting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, "Repushing of gem versions is not allowed.")
	}))
	defer srv.Close()
	file := filepath.Join(t.TempDir(), "acme.gem")
	_ = os.WriteFile(file, []byte("gem"), 0o644)
	if err := (Pusher{Host: srv.URL, Key: "k", Client: srv.Client()}).Push(Gem{File: file}); !errors.Is(err, ErrExists) {
		t.Errorf("err = %v", err)
	}
}

func TestPusherNeedsAKey(t *testing.T) {
	t.Setenv("GEM_HOST_API_KEY", "")
	if _, err := PusherFromEnvironment(); err == nil {
		t.Error("no key, no pusher")
	}
	t.Setenv("GEM_HOST_API_KEY", "k")
	if p, err := PusherFromEnvironment(); err != nil || p.Host != DefaultHost {
		t.Errorf("pusher = %+v, %v", p, err)
	}
}
