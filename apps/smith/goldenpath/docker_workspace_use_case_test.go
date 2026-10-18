//go:build e2e

package goldenpath

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// startRegistry runs a throwaway registry on a free port and returns its
// address (127.0.0.1:port); it is removed when the test ends.
func startRegistry(t *testing.T) string {
	t.Helper()
	name := fmt.Sprintf("smith-e2e-registry-%d", freePort(t))
	if out, err := exec.Command("docker", "run", "-d", "--rm", "--name", name, "-p", "127.0.0.1::5000", "registry:2").CombinedOutput(); err != nil {
		t.Fatalf("starting a registry: %v\n%s", err, out)
	}
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", name).Run() })
	port, err := exec.Command("docker", "port", name, "5000").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(strings.Split(string(port), "\n")[0])
}

// registryTags lists the tags a registry has of an image.
func registryTags(t *testing.T, address, image string) string {
	t.Helper()
	resp, err := http.Get("http://" + address + "/v2/" + image + "/tags/list")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	tags, _ := io.ReadAll(resp.Body)
	return string(tags)
}

// TestDockerWorkspace ships a Go app as a container image with the docker
// convention: the Dockerfile written when the app is added, the image
// task, and a release that builds the image once and pushes the versioned
// image to each of two registries.
func TestDockerWorkspace(t *testing.T) {
	out, err := exec.Command("docker", "version", "--format", "{{.Server.Os}}").Output()
	if err != nil || strings.TrimSpace(string(out)) != "linux" {
		t.Skip("no Docker daemon running Linux containers")
	}
	// Two throwaway registries, on free ports.
	address, mirror := startRegistry(t), startRegistry(t)
	t.Cleanup(func() {
		_ = exec.Command("docker", "image", "rm", "-f", address+"/smith-e2e/api:local", address+"/smith-e2e/api:latest").Run()
	})

	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-docker", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-docker")
	w.write("smith.toml", w.read("smith.toml")+"\n[conventions.docker]\nregistries = [\""+address+"/smith-e2e\", \""+mirror+"/smith-e2e-mirror\"]\n")
	// Settings that contradict each other are refused, naming the setting.
	w.write("smith.toml", w.read("smith.toml")+"registry = \""+address+"/other\"\n")
	if out := w.smith(false, "doctor"); !strings.Contains(out, "sets both registry and registries") {
		t.Fatalf("registry and registries together:\n%s", out)
	}
	w.write("smith.toml", strings.Replace(w.read("smith.toml"), "registry = \""+address+"/other\"\n", "", 1))
	w.smith(true, "add", "app", "api", "--lang", "go")
	if dockerfile := w.read("apps/api/Dockerfile"); !strings.Contains(dockerfile, "go build -trimpath -o /out/app ./apps/api") {
		t.Fatalf("the convention writes a new Go app's Dockerfile:\n%s", dockerfile)
	}
	if _, err := os.Stat(filepath.Join(w.dir, ".dockerignore")); err != nil {
		t.Fatal(".dockerignore keeps the context small")
	}
	if out := w.smith(true, "run", "image", "-p", "api"); !strings.Contains(out, "api:image ✔") {
		t.Fatalf("image:\n%s", out)
	}
	// verify and the image task build the image under the name release
	// pushes, tagged for local use, never with the manifest's version.
	if out, err := exec.Command("docker", "image", "inspect", address+"/smith-e2e/api:local").CombinedOutput(); err != nil {
		t.Fatalf("the image task names the first registry's image: %v\n%s", err, out)
	}
	// Registries on this machine take no login, so the pipeline asks for no
	// container credentials.
	if ci := w.read(".github/workflows/ci.yml"); strings.Contains(ci, "CONTAINER_") {
		t.Fatalf("ci.yml asks for credentials a local registry does not need:\n%s", ci)
	}

	// An app with a Dockerfile releases: its image is pushed with the version.
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: the api")
	if out := w.smith(true, "release", "--dry-run"); !strings.Contains(out, "would push "+address+"/smith-e2e/api:0.1.0") || !strings.Contains(out, mirror+"/smith-e2e-mirror/api:0.1.0") {
		t.Fatalf("the dry run names every image it would push:\n%s", out)
	}
	if out := w.smith(true, "release", "--yes"); !strings.Contains(out, "docker published 0.1.0") {
		t.Fatalf("release:\n%s", out)
	}
	for registry, image := range map[string]string{address: "smith-e2e/api", mirror: "smith-e2e-mirror/api"} {
		if tags := registryTags(t, registry, image); !strings.Contains(tags, `"0.1.0"`) || !strings.Contains(tags, `"latest"`) {
			t.Fatalf("%s has the release's tags: %s", registry, tags)
		}
	}
}

// TestDockerNodeApp ships a Node app that uses a workspace package as an
// image: the image task prunes the app's build (a production package.json
// and lockfile, the package copied into workspace_modules/) and the
// Dockerfile the convention writes installs from it, away from the
// workspace.
func TestDockerNodeApp(t *testing.T) {
	out, err := exec.Command("docker", "version", "--format", "{{.Server.Os}}").Output()
	if err != nil || strings.TrimSpace(string(out)) != "linux" {
		t.Skip("no Docker daemon running Linux containers")
	}
	t.Cleanup(func() { _ = exec.Command("docker", "image", "rm", "-f", "svc:local", "svc:latest").Run() })
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "smith-e2e-node-image", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-node-image")
	w.write("smith.toml", w.read("smith.toml")+"\n[conventions.docker]\n")
	w.smith(true, "add", "package", "greetkit", "--lang", "node")
	w.smith(true, "add", "app", "svc", "--lang", "node")
	var kit struct{ Name string }
	if err := json.Unmarshal([]byte(w.read("packages/greetkit/package.json")), &kit); err != nil {
		t.Fatal(err)
	}
	svc := w.read("apps/svc/package.json")
	w.write("apps/svc/package.json", strings.Replace(svc, `"scripts": {`, `"dependencies": { "`+kit.Name+`": "*" },`+"\n  \"scripts\": {", 1))
	w.write("apps/svc/src/main.ts", "import { greet } from '"+kit.Name+"'\n\nprocess.stdout.write(greet('image').message + '\\n')\n")
	run(t, w.dir, "npm", "install", "--no-audit", "--no-fund")
	if dockerfile := w.read("apps/svc/Dockerfile"); !strings.Contains(dockerfile, "COPY dist/apps/svc/workspace_modules ./workspace_modules") {
		t.Fatalf("the convention writes a Node app's Dockerfile from its pruned build:\n%s", dockerfile)
	}
	if out := w.smith(true, "run", "image", "-p", "svc"); !strings.Contains(out, "svc:prune ✔") || !strings.Contains(out, "svc:image ✔") {
		t.Fatalf("image (after prune):\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(w.dir, "dist", "apps", "svc", "workspace_modules", filepath.FromSlash(kit.Name), "package.json")); err != nil {
		t.Fatalf("prune copies the package the app uses: %v", err)
	}
	got, err := exec.Command("docker", "run", "--rm", "svc:latest").CombinedOutput()
	if err != nil || !strings.Contains(string(got), "image") {
		t.Fatalf("the image runs the app with its workspace package: %v\n%s", err, got)
	}
}
