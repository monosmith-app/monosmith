package releasing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestPublishUploadsOnlyThisVersion(t *testing.T) {
	ws := t.TempDir()
	dist := filepath.Join(ws, "dist", "packages", "sdk")
	_ = os.MkdirAll(dist, 0o755)
	for _, f := range []string{"acme_sdk-1.0.0-py3-none-any.whl", "acme_sdk-1.0.0.tar.gz", "acme_sdk-0.9.0.tar.gz"} {
		_ = os.WriteFile(filepath.Join(dist, f), nil, 0o644)
	}
	httpGet = func(string) (string, error) { return "", nil }
	var argv []string
	Twine = func(_ string, a []string, _ []string) (string, error) {
		argv = a
		return "Uploading acme_sdk-1.0.0\n", nil
	}
	res, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk"}, Version: "1.0.0"})
	if err != nil || !res.Published {
		t.Fatalf("res %+v err %v", res, err)
	}
	joined := strings.Join(argv, " ")
	if strings.Contains(joined, "0.9.0") || strings.Contains(joined, "--skip-existing") || strings.Count(joined, "1.0.0") != 2 {
		t.Fatalf("argv %v", argv)
	}

	httpGet = func(url string) (string, error) {
		if url != "https://pypi.org/simple/acme-sdk/" {
			t.Fatalf("index url %s", url)
		}
		return `<a href="x">acme_sdk-1.0.0-py3-none-any.whl</a><a>acme_sdk-1.0.0.tar.gz</a>`, nil
	}
	Twine = func(string, []string, []string) (string, error) {
		t.Fatal("nothing must be uploaded when the index has every file")
		return "", nil
	}
	res, _ = Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk", Exports: []string{"acme_sdk"}}, Version: "1.0.0"})
	if res.Published || res.Skipped != "already published" {
		t.Fatalf("res %+v", res)
	}
	if _, err := Publish(ws, protocol.PublishParams{Project: protocol.Project{Root: "packages/sdk"}, Version: "2.0.0"}); err == nil {
		t.Fatal("publishing without a built distribution must fail")
	}
}

func TestSimpleIndexURL(t *testing.T) {
	for in, want := range map[string]string{
		"https://upload.pypi.org/legacy/":                         "https://pypi.org/simple/",
		"https://test.pypi.org/legacy/":                           "https://test.pypi.org/simple/",
		"http://localhost:8090/":                                  "http://localhost:8090/simple/",
		"https://pkgs.dev.azure.com/o/p/_packaging/f/pypi/upload": "https://pkgs.dev.azure.com/o/p/_packaging/f/pypi/simple/",
	} {
		if got := SimpleIndexURL(in); got != want {
			t.Errorf("%s: got %s want %s", in, got, want)
		}
	}
	if normalise("Acme_SDK.x") != "acme-sdk-x" {
		t.Fatal("PEP 503 normalisation")
	}
}
