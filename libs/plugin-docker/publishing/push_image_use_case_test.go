package publishing

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fakeDocker replaces Docker and Inspect for the test.
type fakeDocker struct {
	ran       []string          // every docker command, in order
	passwords map[string]string // the password given to each docker login, by host
	has       map[string]bool   // tags a registry already has (manifest inspect succeeds)
	failOn    string            // a command (prefix) that fails
	inspected []string
}

func recordDocker(t *testing.T) *fakeDocker {
	t.Helper()
	realDocker, realInspect := Docker, Inspect
	t.Cleanup(func() { Docker, Inspect = realDocker, realInspect })
	f := &fakeDocker{passwords: map[string]string{}, has: map[string]bool{}}
	Docker = func(_ string, argv []string, stdin string) error {
		cmd := strings.Join(argv, " ")
		f.ran = append(f.ran, cmd)
		if len(argv) > 2 && argv[1] == "login" && stdin != "" {
			f.passwords[argv[2]] = stdin
		}
		if f.failOn != "" && strings.HasPrefix(cmd, f.failOn) {
			return errors.New("denied: requested access to the resource is denied")
		}
		return nil
	}
	Inspect = func(_ string, argv []string) error {
		f.inspected = append(f.inspected, strings.Join(argv, " "))
		if f.has[argv[len(argv)-1]] {
			return nil
		}
		return errors.New("no such manifest")
	}
	return f
}

func shipping(t *testing.T, settings map[string]any) (string, protocol.Project) {
	t.Helper()
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "api", "Dockerfile"), nil, 0o644)
	api := protocol.Project{Name: "api", Root: "apps/api", Kind: protocol.ProjectApp}
	if settings != nil {
		api.Settings = map[string]map[string]any{"docker": settings}
	}
	for _, env := range []string{"CONTAINER_REGISTRY", "CONTAINER_USERNAME", "CONTAINER_PASSWORD", "GITHUB_TOKEN", "GITHUB_ACTOR", "DOCKERHUB_USERNAME", "DOCKERHUB_TOKEN"} {
		t.Setenv(env, "")
	}
	return ws, api
}

func TestPublishWithoutARegistryBuildsButDoesNotPush(t *testing.T) {
	ws, api := shipping(t, nil)
	f := recordDocker(t)
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || res.Published || !strings.Contains(res.Skipped, "[conventions.docker] registry") || len(f.ran) != 0 {
		t.Fatalf("%+v %v %v", res, err, f.ran)
	}
	_, off := shipping(t, map[string]any{"registry": "ghcr.io/acme", "push": false})
	if res, err := Publish(ws, protocol.PublishParams{Project: off, Version: "1.0.0"}); err != nil || !strings.Contains(res.Skipped, "push = false") {
		t.Fatalf("push = false: %+v %v", res, err)
	}
}

func TestPublishToGHCRLogsInWithTheJobsToken(t *testing.T) {
	ws, api := shipping(t, map[string]any{"registry": "ghcr.io/acme/", "image": "acme-api"})
	f := recordDocker(t)
	if res, _ := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0", DryRun: true}); res.Skipped != "would push ghcr.io/acme/acme-api:1.0.0, ghcr.io/acme/acme-api:latest" || len(f.ran) != 0 {
		t.Fatalf("a dry run names the images and runs nothing: %+v %v", res, f.ran)
	}
	t.Setenv("GITHUB_TOKEN", "ghs_token")
	t.Setenv("GITHUB_ACTOR", "octocat")
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || !res.Published || res.Registry != "ghcr.io/acme" {
		t.Fatalf("%+v %v", res, err)
	}
	want := []string{
		"docker login ghcr.io -u octocat --password-stdin",
		"docker build -f apps/api/Dockerfile -t ghcr.io/acme/acme-api:1.0.0 -t ghcr.io/acme/acme-api:latest .",
		"docker push ghcr.io/acme/acme-api:latest",
		"docker push ghcr.io/acme/acme-api:1.0.0",
	}
	if strings.Join(f.ran, "\n") != strings.Join(want, "\n") || f.passwords["ghcr.io"] != "ghs_token" {
		t.Fatalf("ran:\n%s", strings.Join(f.ran, "\n"))
	}
}

func TestPublishToAnotherRegistryLogsInWithContainerCredentials(t *testing.T) {
	ws, api := shipping(t, map[string]any{"registry": "registry.example.com/team"})
	f := recordDocker(t)
	t.Setenv("CONTAINER_USERNAME", "bot")
	t.Setenv("CONTAINER_PASSWORD", "s3cret")
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"}); err != nil {
		t.Fatal(err)
	}
	if f.ran[0] != "docker login registry.example.com -u bot --password-stdin" || f.passwords["registry.example.com"] != "s3cret" {
		t.Fatalf("ran:\n%s", strings.Join(f.ran, "\n"))
	}
}

func TestPublishStillReadsContainerRegistry(t *testing.T) {
	ws, api := shipping(t, nil)
	f := recordDocker(t)
	t.Setenv("CONTAINER_REGISTRY", "127.0.0.1:5000/e2e/")
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || !res.Published || res.Registry != "127.0.0.1:5000/e2e" || f.ran[len(f.ran)-1] != "docker push 127.0.0.1:5000/e2e/api:1.0.0" {
		t.Fatalf("%+v %v %v", res, err, f.ran)
	}
	if res, _ := Publish(ws, protocol.PublishParams{Project: protocol.Project{Name: "web", Root: "apps/web", Kind: protocol.ProjectApp}}); res.Skipped != "no Dockerfile" {
		t.Errorf("%+v", res)
	}
}

// both is ghcr.io and Docker Hub, the OpenCraw case.
var both = map[string]any{"registries": []any{"ghcr.io/open-craw", "docker.io/opencraw"}, "image": "opencraw-api"}

func bothCredentials(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_TOKEN", "ghs_token")
	t.Setenv("GITHUB_ACTOR", "octocat")
	t.Setenv("DOCKERHUB_USERNAME", "opencraw")
	t.Setenv("DOCKERHUB_TOKEN", "dckr_pat")
}

func TestPublishBuildsOnceAndPushesToEveryRegistry(t *testing.T) {
	ws, api := shipping(t, both)
	f := recordDocker(t)
	if res, _ := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0", DryRun: true}); res.Skipped !=
		"would push ghcr.io/open-craw/opencraw-api:1.0.0, ghcr.io/open-craw/opencraw-api:latest, docker.io/opencraw/opencraw-api:1.0.0, docker.io/opencraw/opencraw-api:latest" || len(f.ran) != 0 {
		t.Fatalf("the dry run lists every reference: %+v %v", res, f.ran)
	}
	bothCredentials(t)
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || !res.Published || res.Registry != "ghcr.io/open-craw, docker.io/opencraw" {
		t.Fatalf("%+v %v", res, err)
	}
	want := []string{
		"docker login ghcr.io -u octocat --password-stdin",
		"docker login docker.io -u opencraw --password-stdin",
		"docker build -f apps/api/Dockerfile -t ghcr.io/open-craw/opencraw-api:1.0.0 -t ghcr.io/open-craw/opencraw-api:latest -t docker.io/opencraw/opencraw-api:1.0.0 -t docker.io/opencraw/opencraw-api:latest .",
		"docker push ghcr.io/open-craw/opencraw-api:latest",
		"docker push ghcr.io/open-craw/opencraw-api:1.0.0",
		"docker push docker.io/opencraw/opencraw-api:latest",
		"docker push docker.io/opencraw/opencraw-api:1.0.0",
	}
	if strings.Join(f.ran, "\n") != strings.Join(want, "\n") {
		t.Fatalf("the image is built once, then pushed to each registry:\n%s", strings.Join(f.ran, "\n"))
	}
	if f.passwords["ghcr.io"] != "ghs_token" || f.passwords["docker.io"] != "dckr_pat" {
		t.Errorf("each registry gets its own password: %v", f.passwords)
	}
}

func TestAFailedPushNamesItsRegistryAndAReRunResumes(t *testing.T) {
	ws, api := shipping(t, both)
	bothCredentials(t)
	f := recordDocker(t)
	f.failOn = "docker push docker.io/opencraw/opencraw-api:latest"
	_, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err == nil || !strings.Contains(err.Error(), "docker.io/opencraw: denied") || strings.Contains(err.Error(), "ghcr.io/open-craw: denied") ||
		!strings.Contains(err.Error(), "ghcr.io/open-craw already have 1.0.0") {
		t.Fatalf("the error names the registry that failed and the one that did not: %v", err)
	}
	if got := f.ran[len(f.ran)-2:]; got[0] != "docker push ghcr.io/open-craw/opencraw-api:1.0.0" {
		t.Errorf("ghcr.io was still pushed to: %v", f.ran)
	}

	// The re-run: ghcr.io has the version, docker.io does not.
	f = recordDocker(t)
	f.has["ghcr.io/open-craw/opencraw-api:1.0.0"] = true
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || !res.Published || res.Registry != "docker.io/opencraw" {
		t.Fatalf("%+v %v", res, err)
	}
	for _, cmd := range f.ran {
		if strings.Contains(cmd, "ghcr.io/open-craw/opencraw-api") {
			t.Errorf("the re-run pushed to a registry that has the version: %s", cmd)
		}
	}
	if got := strings.Join(f.ran, "\n"); !strings.Contains(got, "docker build -f apps/api/Dockerfile -t docker.io/opencraw/opencraw-api:1.0.0 -t docker.io/opencraw/opencraw-api:latest .") {
		t.Errorf("only the missing registry's tags are built:\n%s", got)
	}
}

func TestPublishLeavesRegistriesThatHaveTheVersion(t *testing.T) {
	ws, api := shipping(t, both)
	bothCredentials(t)
	f := recordDocker(t)
	f.has["ghcr.io/open-craw/opencraw-api:1.0.0"] = true
	f.has["docker.io/opencraw/opencraw-api:1.0.0"] = true
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || res.Published || !strings.Contains(res.Skipped, "already published to ghcr.io/open-craw, docker.io/opencraw") {
		t.Fatalf("%+v %v", res, err)
	}
	for _, cmd := range f.ran {
		if strings.HasPrefix(cmd, "docker build") || strings.HasPrefix(cmd, "docker push") {
			t.Errorf("nothing to build or push: %s", cmd)
		}
	}
}

func TestALocalRegistryIsInspectedOverHTTP(t *testing.T) {
	ws, api := shipping(t, map[string]any{"registries": []any{"127.0.0.1:5000/e2e", "localhost:5001/e2e"}})
	f := recordDocker(t)
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"}); err != nil {
		t.Fatal(err)
	}
	if len(f.inspected) != 2 || f.inspected[0] != "docker manifest inspect --insecure 127.0.0.1:5000/e2e/api:1.0.0" {
		t.Errorf("inspected %v", f.inspected)
	}
	for _, cmd := range f.ran {
		if strings.HasPrefix(cmd, "docker login") {
			t.Errorf("a registry on this machine takes no login: %s", cmd)
		}
	}
}

func TestContradictorySettingsFailThePublish(t *testing.T) {
	ws, api := shipping(t, map[string]any{"registry": "ghcr.io/a", "registries": []any{"docker.io/b"}})
	f := recordDocker(t)
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"}); err == nil || len(f.ran) != 0 {
		t.Fatalf("%v %v", err, f.ran)
	}
}
