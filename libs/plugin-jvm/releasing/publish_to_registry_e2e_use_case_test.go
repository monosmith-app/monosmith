//go:build e2e

package releasing

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestPublishDeploysThroughMavenToAGitHubPackagesStyleRepository runs the
// real Publish with the real mvn against a local repository that wants
// basic auth, as GitHub Packages does: the job's token, no Maven secrets.
func TestPublishDeploysThroughMavenToAGitHubPackagesStyleRepository(t *testing.T) {
	if _, err := exec.LookPath("mvn"); err != nil {
		t.Skip("mvn is not installed")
	}
	var mu sync.Mutex
	stored := map[string]string{}
	authorized := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, _ := r.BasicAuth()
		mu.Lock()
		defer mu.Unlock()
		if user != "octocat" || password != "job-token" {
			authorized = false
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			stored[r.URL.Path] = string(body)
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	ws := t.TempDir()
	pom := `<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>dev.acme</groupId>
  <artifactId>acme-core</artifactId>
  <version>1.0.0</version>
</project>
`
	if err := os.WriteFile(filepath.Join(ws, "pom.xml"), []byte(pom), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_TOKEN", "job-token")
	t.Setenv("GITHUB_ACTOR", "octocat")
	t.Setenv("MAVEN_USERNAME", "")
	t.Setenv("MAVEN_PASSWORD", "")
	p := protocol.Project{
		Name: "core", Root: ".", Kind: protocol.ProjectPackage, Exports: []string{"dev.acme:acme-core"},
		Registries: &protocol.Registries{Publish: &protocol.Registry{Name: "github", Auth: protocol.AuthGitHubPackages, Maven: srv.URL + "/acme/tools"}},
	}
	res, err := Publish(ws, protocol.PublishParams{Project: p, Version: "1.0.0"})
	if err != nil || !res.Published {
		t.Fatalf("%+v %v", res, err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !authorized {
		t.Error("the repository got a request without the job's credentials")
	}
	if _, ok := stored["/acme/tools/dev/acme/acme-core/1.0.0/acme-core-1.0.0.pom"]; !ok {
		var paths []string
		for k := range stored {
			paths = append(paths, k)
		}
		t.Fatalf("the pom was not deployed: %s", strings.Join(paths, ", "))
	}
}
