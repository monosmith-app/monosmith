package releasing

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// SimpleIndexURL returns the PEP 503 "simple" index matching an upload URL:
// upload.pypi.org/legacy/ → pypi.org/simple/, test.pypi.org likewise, and any
// other index <url>/simple/. SMITH_PYPI_SIMPLE_URL overrides it.
func SimpleIndexURL(uploadURL string) string {
	if env := os.Getenv("SMITH_PYPI_SIMPLE_URL"); env != "" {
		return strings.TrimSuffix(env, "/") + "/"
	}
	u := strings.TrimSuffix(uploadURL, "/")
	switch {
	case strings.Contains(u, "upload.pypi.org"):
		return "https://pypi.org/simple/"
	case strings.Contains(u, "test.pypi.org"):
		return "https://test.pypi.org/simple/"
	}
	// Azure Artifacts names its upload URL .../pypi/upload/ beside the
	// .../pypi/simple/ index.
	if strings.HasSuffix(u, "/upload") {
		return strings.TrimSuffix(u, "/upload") + "/simple/"
	}
	u = strings.TrimSuffix(u, "/legacy")
	return u + "/simple/"
}

// httpGet fetches a page; tests replace it.
var httpGet = func(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	return string(body), err
}

// Uploaded reports which file names the index already has for a project.
// An unreachable index reports nothing, so the upload is attempted.
func Uploaded(simpleURL, project string, files []string) map[string]bool {
	page, err := httpGet(simpleURL + normalise(project) + "/")
	out := map[string]bool{}
	if err != nil {
		return out
	}
	for _, f := range files {
		if strings.Contains(page, f) {
			out[f] = true
		}
	}
	return out
}

// normalise is the PEP 503 project name normalisation.
func normalise(name string) string {
	name = strings.ToLower(name)
	var b strings.Builder
	dash := false
	for _, r := range name {
		if r == '-' || r == '_' || r == '.' {
			if !dash {
				b.WriteByte('-')
			}
			dash = true
			continue
		}
		dash = false
		b.WriteRune(r)
	}
	return b.String()
}
