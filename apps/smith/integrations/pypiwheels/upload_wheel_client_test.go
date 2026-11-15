package pypiwheels

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadSendsTheWheelAndItsMetadata(t *testing.T) {
	var form map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		if user != "__token__" || pass != "pypi-secret" {
			t.Errorf("auth %s:%s", user, pass)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		form = map[string]string{}
		for k, v := range r.MultipartForm.Value {
			form[k] = v[0]
		}
		if f := r.MultipartForm.File["content"]; len(f) != 1 || f[0].Filename != "acme-5.0.0-py3-none-win_amd64.whl" {
			t.Errorf("content = %v", f)
		}
	}))
	defer server.Close()
	file := filepath.Join(t.TempDir(), "acme-5.0.0-py3-none-win_amd64.whl")
	_ = os.WriteFile(file, []byte("wheel"), 0o644)
	u := Uploader{URL: server.URL, Username: "__token__", Password: "pypi-secret", Client: server.Client()}
	if err := u.Upload(Wheel{File: file, Metadata: Metadata{Name: "acme", Version: "5.0.0", Summary: "Does things"}}); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]string{":action": "file_upload", "name": "acme", "version": "5.0.0", "filetype": "bdist_wheel", "pyversion": "py3", "summary": "Does things"} {
		if form[k] != want {
			t.Errorf("%s = %q, want %q", k, form[k], want)
		}
	}
	if len(form["sha256_digest"]) != 64 {
		t.Errorf("sha256_digest = %q", form["sha256_digest"])
	}
}

func TestUploadOfAFileTheIndexHasIsErrExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "400 File already exists ('acme-5.0.0-py3-none-win_amd64.whl')", http.StatusBadRequest)
	}))
	defer server.Close()
	file := filepath.Join(t.TempDir(), "acme.whl")
	_ = os.WriteFile(file, []byte("wheel"), 0o644)
	u := Uploader{URL: server.URL, Client: server.Client()}
	if err := u.Upload(Wheel{File: file}); !errors.Is(err, ErrExists) {
		t.Errorf("err = %v", err)
	}
}

func TestUploaderFromEnvironmentNeedsAToken(t *testing.T) {
	t.Setenv("PYPI_TOKEN", "")
	t.Setenv("TWINE_PASSWORD", "")
	if _, err := UploaderFromEnvironment(); err == nil {
		t.Error("no credentials accepted")
	}
	t.Setenv("PYPI_TOKEN", "pypi-secret")
	t.Setenv("TWINE_REPOSITORY_URL", "")
	u, err := UploaderFromEnvironment()
	if err != nil || u.Username != "__token__" || u.Password != "pypi-secret" || u.URL != DefaultUploadURL {
		t.Errorf("uploader %+v, %v", u, err)
	}
}
