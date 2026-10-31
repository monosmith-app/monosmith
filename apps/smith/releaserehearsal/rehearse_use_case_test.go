package releaserehearsal

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func envOf(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func statuses(r Report) map[string]string {
	out := map[string]string{}
	for _, res := range r.Results {
		out[res.Check] = res.Status
	}
	return out
}

func run(t *testing.T, target Target, env map[string]string) Report {
	t.Helper()
	return Rehearse([]Target{target}, NewProbe(), envOf(env))
}

func TestNPMBearerAccepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/-/whoami" && r.Header.Get("Authorization") == "Bearer s3cret" {
			_, _ = w.Write([]byte(`{"username":"bob"}`))
			return
		}
		if r.URL.Path == "/-/whoami" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()
	rep := run(t, Target{Project: "core", Ecosystem: NPM, URL: srv.URL, Names: []string{"core"}}, map[string]string{"NODE_AUTH_TOKEN": "s3cret"})
	if rep.Failed != 0 || statuses(rep)[CheckAccess] != Pass {
		t.Fatalf("want all passing: %+v", rep.Results)
	}
	if text := RenderText(rep, 1, true); strings.Contains(text, "s3cret") || !strings.Contains(text, "bob") {
		t.Fatalf("the text must name the user and never the token:\n%s", text)
	}

	rep = run(t, Target{Project: "core", Ecosystem: NPM, URL: srv.URL, Names: []string{"core"}}, map[string]string{"NODE_AUTH_TOKEN": "wrong"})
	if rep.Failed != 1 || !strings.Contains(rep.Results[len(rep.Results)-1].Fix, "NODE_AUTH_TOKEN") {
		t.Fatalf("a refused token fails and names the variable: %+v", rep.Results)
	}
}

func TestNPMBasicFeedWithoutWhoami(t *testing.T) {
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("smith:pipeline"))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != want {
			w.Header().Set("WWW-Authenticate", `Basic realm="feed"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNotFound) // no whoami, and the package is new
	}))
	defer srv.Close()
	target := Target{Project: "core", Ecosystem: NPM, URL: srv.URL, Auth: "azure-artifacts", Names: []string{"@acme/core"}}
	rep := run(t, target, map[string]string{"TF_BUILD": "True", "SYSTEM_ACCESSTOKEN": "pipeline"})
	if rep.Failed != 0 || statuses(rep)[CheckIdentity] != Pass {
		t.Fatalf("want pass: %+v", rep.Results)
	}
	if !strings.Contains(rep.Results[0].Message, "Basic") {
		t.Fatalf("the challenge must be reported: %q", rep.Results[0].Message)
	}
	rep = run(t, target, map[string]string{"TF_BUILD": "True"})
	last := rep.Results[len(rep.Results)-1]
	if rep.Failed != 1 || !strings.Contains(last.Fix, "System.AccessToken") {
		t.Fatalf("a missing pipeline token fails with the mapping: %+v", rep.Results)
	}
	rep = run(t, target, map[string]string{})
	if rep.Failed != 0 || rep.Warned != 1 {
		t.Fatalf("Azure's token cannot exist outside Azure: a warning, not a failure: %+v", rep.Results)
	}
}

func TestNPMNpmrcTokenFromVariable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer fromrc" {
			_, _ = w.Write([]byte("{}"))
			return
		}
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	target := Target{Project: "core", Ecosystem: NPM, URL: srv.URL, Names: []string{"core"}, NpmrcToken: "fromrc", NpmrcVar: "FEED"}
	if rep := run(t, target, nil); rep.Failed != 0 {
		t.Fatalf("an .npmrc token is a credential: %+v", rep.Results)
	}
}

func TestPyPIEmptyUpload(t *testing.T) {
	var gotFile bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		user, pass, _ := r.BasicAuth()
		_ = r.ParseMultipartForm(1 << 20)
		gotFile = r.MultipartForm != nil && len(r.MultipartForm.File) > 0
		if user != "__token__" || pass != "pypi-ok" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusBadRequest) // no file
	}))
	defer srv.Close()
	target := Target{Project: "lib", Ecosystem: PyPI, URL: srv.URL, Names: []string{"lib"}}
	if rep := run(t, target, map[string]string{"PYPI_TOKEN": "pypi-ok"}); rep.Failed != 0 || gotFile {
		t.Fatalf("want pass and no file sent: %+v", rep.Results)
	}
	rep := run(t, target, map[string]string{"PYPI_TOKEN": "bad"})
	if rep.Failed != 1 || !strings.Contains(rep.Results[len(rep.Results)-1].Fix, "PYPI_TOKEN") {
		t.Fatalf("want a failure naming PYPI_TOKEN: %+v", rep.Results)
	}
}

func TestNuGetPrivateFeed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pass, ok := r.BasicAuth(); !ok || pass != "pat" {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"feed\"")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"version":"3.0.0"}`))
	}))
	defer srv.Close()
	target := Target{Project: "sdk", Ecosystem: NuGet, URL: srv.URL + "/index.json", Auth: "token", TokenEnv: "FEED_PAT", Names: []string{"Acme.Sdk"}}
	if rep := run(t, target, map[string]string{"FEED_PAT": "pat"}); rep.Failed != 0 || rep.Passed != 3 {
		t.Fatalf("want pass: %+v", rep.Results)
	}
	rep := run(t, target, map[string]string{"FEED_PAT": "nope"})
	if rep.Failed != 1 {
		t.Fatalf("want a refusal: %+v", rep.Results)
	}
	if rep := run(t, target, nil); rep.Failed != 1 || !strings.Contains(rep.Results[1].Message, "FEED_PAT") {
		t.Fatalf("a missing variable is named: %+v", rep.Results)
	}
}

func TestNuGetOrgKeyIsOnlyNoted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{}")) }))
	defer srv.Close()
	rep := run(t, Target{Project: "sdk", Ecosystem: NuGet, URL: srv.URL, Names: []string{"Acme.Sdk"}}, map[string]string{"NUGET_API_KEY": "k"})
	if rep.Failed != 0 || !strings.Contains(rep.Results[2].Message, "without pushing") {
		t.Fatalf("an unverifiable key is said, not failed: %+v", rep.Results)
	}
}

func TestCratesToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/me" && r.Header.Get("Authorization") != "cio" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()
	target := Target{Project: "crate", Ecosystem: Crates, URL: srv.URL, Names: []string{"crate"}}
	if rep := run(t, target, map[string]string{"CARGO_REGISTRY_TOKEN": "cio"}); rep.Failed != 0 {
		t.Fatalf("want pass: %+v", rep.Results)
	}
	if rep := run(t, target, map[string]string{"CARGO_REGISTRY_TOKEN": "bad"}); rep.Failed != 1 {
		t.Fatalf("want a refusal: %+v", rep.Results)
	}
}

func TestUnreachableAndOIDC(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()
	rep := run(t, Target{Project: "core", Ecosystem: NPM, URL: url, Names: []string{"core"}}, nil)
	if rep.Failed != 1 || len(rep.Results) != 1 || rep.Results[0].Fix == "" {
		t.Fatalf("an unreachable registry fails once with a fix: %+v", rep.Results)
	}
	up := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer up.Close()
	rep = run(t, Target{Project: "core", Ecosystem: NPM, URL: up.URL, Names: []string{"core"}}, map[string]string{"GITHUB_ACTIONS": "true", "ACTIONS_ID_TOKEN_REQUEST_URL": "https://x"})
	if rep.Failed != 0 || rep.Warned != 1 {
		t.Fatalf("OIDC stands in for a token as a warning: %+v", rep.Results)
	}
	rep = run(t, Target{Project: "core", Ecosystem: NPM, URL: up.URL, Names: []string{"core"}}, map[string]string{"GITHUB_ACTIONS": "true"})
	if rep.Failed != 1 || !strings.Contains(rep.Results[1].Fix, "id-token: write") {
		t.Fatalf("a GitHub job with neither fails with both ways out: %+v", rep.Results)
	}
}
