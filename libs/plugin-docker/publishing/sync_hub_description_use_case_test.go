package publishing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// fakeHub stands in for Docker Hub's API and records what it was asked.
type fakeHub struct {
	requests     []string // "METHOD path"
	loginBody    map[string]string
	description  string
	authorized   string
	patchStatus  int
	loginStatus  int
	descriptions int
}

func startHub(t *testing.T) *fakeHub {
	t.Helper()
	hub := &fakeHub{patchStatus: http.StatusOK, loginStatus: http.StatusOK}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.requests = append(hub.requests, r.Method+" "+r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		switch r.Method {
		case http.MethodPost:
			_ = json.Unmarshal(raw, &hub.loginBody)
			w.WriteHeader(hub.loginStatus)
			_, _ = w.Write([]byte(`{"token":"jwt-token"}`))
		case http.MethodPatch:
			var body map[string]string
			_ = json.Unmarshal(raw, &body)
			hub.description, hub.authorized = body["full_description"], r.Header.Get("Authorization")
			hub.descriptions++
			w.WriteHeader(hub.patchStatus)
			_, _ = w.Write([]byte(`{"detail":"answer"}`))
		}
	}))
	realURL, realClient := hubURL, hubClient
	hubURL, hubClient = server.URL, server.Client()
	t.Cleanup(func() { server.Close(); hubURL, hubClient = realURL, realClient })
	return hub
}

// capture collects what the plugin logs, warnings included.
func capture(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	real := toolrun.Log
	toolrun.Log = &buf
	t.Cleanup(func() { toolrun.Log = real })
	return &buf
}

// withReadme writes the workspace's README.md.
func withReadme(t *testing.T, ws, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(ws, "README.md"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hubSettings() map[string]any {
	s := map[string]any{"registries": []any{"ghcr.io/open-craw", "docker.io/opencraw"}, "image": "opencraw-api", "description_from": "README.md"}
	return s
}

func TestThePushSetsTheDockerHubDescriptionFromAFile(t *testing.T) {
	ws, api := shipping(t, hubSettings())
	bothCredentials(t)
	recordDocker(t)
	hub := startHub(t)
	withReadme(t, ws, "# opencraw\n\nThe API.")
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || !res.Published {
		t.Fatalf("%+v %v", res, err)
	}
	if got := strings.Join(hub.requests, ", "); got != "POST /v2/users/login/, PATCH /v2/repositories/opencraw/opencraw-api/" {
		t.Errorf("only the docker.io registry is asked, to sign in and to set the description: %s", got)
	}
	if hub.loginBody["username"] != "opencraw" || hub.loginBody["password"] != "dckr_pat" {
		t.Errorf("signs in with the registry's credentials: %v", hub.loginBody)
	}
	if hub.authorized != "JWT jwt-token" || hub.description != "# opencraw\n\nThe API." {
		t.Errorf("authorization %q, description %q", hub.authorized, hub.description)
	}
}

func TestADescriptionThatFailsIsOnlyAWarning(t *testing.T) {
	ws, api := shipping(t, hubSettings())
	bothCredentials(t)
	recordDocker(t)
	hub := startHub(t)
	hub.patchStatus = http.StatusForbidden
	log := capture(t)
	withReadme(t, ws, "# opencraw")
	res, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"})
	if err != nil || !res.Published {
		t.Fatalf("the images are pushed, so the release goes on: %+v %v", res, err)
	}
	if got := log.String(); !strings.Contains(got, "warning: the Docker Hub description of docker.io/opencraw/opencraw-api was not set") || !strings.Contains(got, "403 Forbidden") || !strings.Contains(got, "Read, Write & Delete") {
		t.Errorf("the warning says what failed and what to try:\n%s", got)
	}

	// A refused sign-in is a warning as well.
	hub.loginStatus = http.StatusUnauthorized
	log.Reset()
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.1"}); err != nil || !strings.Contains(log.String(), "signing in") {
		t.Errorf("%v\n%s", err, log)
	}
}

func TestNoDescriptionIsSetWithoutTheSettingTheCredentialsOrAPush(t *testing.T) {
	ws, api := shipping(t, map[string]any{"registries": []any{"docker.io/opencraw"}})
	bothCredentials(t)
	recordDocker(t)
	hub := startHub(t)
	withReadme(t, ws, "# opencraw")
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"}); err != nil || len(hub.requests) != 0 {
		t.Fatalf("no description_from, no request: %v %v", err, hub.requests)
	}

	// description_from, but the push fails: nothing to describe.
	ws, api = shipping(t, hubSettings())
	bothCredentials(t)
	f := recordDocker(t)
	f.failOn = "docker push"
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"}); err == nil || len(hub.requests) != 0 {
		t.Fatalf("a failed push sets no description: %v %v", err, hub.requests)
	}

	// description_from, but no Docker Hub secrets: a warning, no request.
	ws, api = shipping(t, hubSettings())
	t.Setenv("GITHUB_TOKEN", "ghs_token")
	recordDocker(t)
	log := capture(t)
	withReadme(t, ws, "# opencraw")
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"}); err != nil || len(hub.requests) != 0 || !strings.Contains(log.String(), "DOCKERHUB_USERNAME and DOCKERHUB_TOKEN are not both set") {
		t.Fatalf("%v %v\n%s", err, hub.requests, log)
	}
}

func TestTheDescriptionCanBeSetWithoutAReleaseOrAPush(t *testing.T) {
	ws, api := shipping(t, hubSettings())
	bothCredentials(t)
	hub := startHub(t)
	withReadme(t, ws, "# opencraw\n\nThe API.")
	res, err := SyncDescription(ws, api, false)
	if err != nil || len(res.Failed) != 0 || len(res.Set) != 1 || res.Set[0] != "docker.io/opencraw/opencraw-api" {
		t.Fatalf("%+v %v", res, err)
	}
	if hub.description != "# opencraw\n\nThe API." || hub.authorized != "JWT jwt-token" {
		t.Errorf("authorization %q, description %q", hub.authorized, hub.description)
	}
	// A second call sets it again: the call is idempotent, so there is
	// nothing to remember between runs.
	if _, err := SyncDescription(ws, api, false); err != nil || hub.descriptions != 2 {
		t.Errorf("%v, %d descriptions", err, hub.descriptions)
	}
}

func TestADryRunNamesTheRepositoriesAndCallsNothing(t *testing.T) {
	ws, api := shipping(t, hubSettings())
	bothCredentials(t)
	hub := startHub(t)
	withReadme(t, ws, "# opencraw")
	res, err := SyncDescription(ws, api, true)
	if err != nil || len(res.Set) != 1 || len(hub.requests) != 0 {
		t.Errorf("%+v %v, requests %v", res, err, hub.requests)
	}
}

func TestAFailedDescriptionIsInTheResultAndLogged(t *testing.T) {
	ws, api := shipping(t, hubSettings())
	bothCredentials(t)
	hub := startHub(t)
	hub.patchStatus = http.StatusForbidden
	log := capture(t)
	withReadme(t, ws, "# opencraw")
	res, err := SyncDescription(ws, api, false)
	if err != nil || len(res.Failed) != 1 || !strings.Contains(res.Failed[0], "403 Forbidden") || !strings.Contains(log.String(), "Read, Write & Delete") {
		t.Errorf("%+v %v\n%s", res, err, log)
	}
}

func TestNothingIsSetWithoutADescriptionFileOrADockerHubRegistry(t *testing.T) {
	ws, api := shipping(t, map[string]any{"registries": []any{"docker.io/opencraw"}})
	bothCredentials(t)
	hub := startHub(t)
	if res, err := SyncDescription(ws, api, false); err != nil || res.Skipped != "no description_from" || len(hub.requests) != 0 {
		t.Errorf("%+v %v", res, err)
	}
	ws, api = shipping(t, map[string]any{"registries": []any{"ghcr.io/open-craw"}, "description_from": "README.md"})
	if res, err := SyncDescription(ws, api, false); err != nil || res.Skipped != "no docker.io registry" || len(hub.requests) != 0 {
		t.Errorf("%+v %v", res, err)
	}
}

func TestAMissingOrLongDescriptionFileIsHandled(t *testing.T) {
	ws, api := shipping(t, hubSettings())
	bothCredentials(t)
	recordDocker(t)
	hub := startHub(t)
	log := capture(t)
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.0"}); err != nil || len(hub.requests) != 0 || !strings.Contains(log.String(), "not setting the Docker Hub description") {
		t.Fatalf("a missing README is a warning: %v %v\n%s", err, hub.requests, log)
	}
	withReadme(t, ws, strings.Repeat("é", hubDescriptionLimit))
	log.Reset()
	if _, err := Publish(ws, protocol.PublishParams{Project: api, Version: "1.0.1"}); err != nil {
		t.Fatal(err)
	}
	if len(hub.description) == 0 || len(hub.description) > hubDescriptionLimit || !strings.Contains(log.String(), "the description is cut") {
		t.Errorf("a long README is cut to Docker Hub's limit: %d bytes\n%s", len(hub.description), log)
	}
}
