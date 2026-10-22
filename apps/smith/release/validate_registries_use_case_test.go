package release

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fakeRegistries serves npm, PyPI, NuGet and RubyGems and records every
// request that would publish something.
type fakeRegistries struct {
	server    *httptest.Server
	mu        sync.Mutex
	publishes []string
	// rubyKey is the only RubyGems key accepted.
	rubyKey string
	// rubyPushOnly is a key RubyGems knows but without the index scope.
	rubyPushOnly string
	// npmVersions are the versions npm already has of "acme".
	npmVersions []string
	// npmMaintainers maintain "acme".
	npmMaintainers []string
}

func newFakeRegistries(t *testing.T) *fakeRegistries {
	t.Helper()
	f := &fakeRegistries{rubyKey: "good-ruby-key", npmMaintainers: []string{"ci-bot"}}
	mux := http.NewServeMux()
	mux.HandleFunc("/npm/-/whoami", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer good-npm-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"username":"ci-bot"}`))
	})
	mux.HandleFunc("/npm/acme", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			f.record(r)
		}
		versions := map[string]any{}
		for _, v := range f.npmVersions {
			versions[v] = map[string]any{}
		}
		var maintainers []map[string]string
		for _, m := range f.npmMaintainers {
			maintainers = append(maintainers, map[string]string{"name": m})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"versions": versions, "maintainers": maintainers})
	})
	mux.HandleFunc("/pypi/upload", func(w http.ResponseWriter, r *http.Request) {
		if _, pass, _ := r.BasicAuth(); pass != "good-pypi-token" {
			http.Error(w, "Invalid or non-existent authentication information.", http.StatusForbidden)
			return
		}
		if err := r.ParseMultipartForm(1 << 20); err == nil && r.MultipartForm != nil && len(r.MultipartForm.File) > 0 {
			f.record(r)
		}
		http.Error(w, "Include at least one file", http.StatusBadRequest)
	})
	mux.HandleFunc("/nuget/push", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-NuGet-ApiKey") != "good-nuget-key" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "no package", http.StatusBadRequest)
	})
	mux.HandleFunc("/ruby/api/v1/gems.json", func(w http.ResponseWriter, r *http.Request) {
		if key := r.Header.Get("Authorization"); f.rubyPushOnly != "" && key == f.rubyPushOnly {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		} else if key != f.rubyKey {
			http.Error(w, "Access Denied", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`[]`))
	})
	mux.HandleFunc("/ruby/api/v1/gems", func(w http.ResponseWriter, r *http.Request) { f.record(r) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeRegistries) record(r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.publishes = append(f.publishes, r.Method+" "+r.URL.Path)
}

func (f *fakeRegistries) probe() *RegistryProbe {
	return &RegistryProbe{
		Client:      f.server.Client(),
		NPMRegistry: f.server.URL + "/npm", NPMToken: "good-npm-token",
		PyPIUpload: f.server.URL + "/pypi/upload", PyPIUser: "__token__", PyPIPassword: "good-pypi-token",
		NuGetPush: f.server.URL + "/nuget/push", NuGetKey: "good-nuget-key",
		RubyGemsHost: f.server.URL + "/ruby", RubyGemsKey: "good-ruby-key",
	}
}

func allTargets() []RegistryTarget {
	return []RegistryTarget{
		{Project: "cli", Registry: "npm", Names: []string{"acme"}, Version: "1.2.0"},
		{Project: "cli", Registry: "PyPI", Names: []string{"acme"}, Version: "1.2.0"},
		{Project: "cli", Registry: "NuGet", Names: []string{"acme"}, Version: "1.2.0"},
		{Project: "cli", Registry: "RubyGems", Names: []string{"acme"}, Version: "1.2.0"},
	}
}

func blocking(findings []RegistryFinding) []RegistryFinding {
	var out []RegistryFinding
	for _, f := range findings {
		if f.Blocking {
			out = append(out, f)
		}
	}
	return out
}

func TestAcceptedCredentialsOnEveryRegistryGiveNoBlockingFinding(t *testing.T) {
	f := newFakeRegistries(t)
	findings := ValidateRegistries(allTargets(), f.probe())
	if len(findings) != 0 {
		t.Fatalf("findings: %+v", findings)
	}
	if len(f.publishes) != 0 {
		t.Errorf("the check published: %v", f.publishes)
	}
}

func TestABadCredentialOnTheLastRegistryIsFoundBeforeAnythingIsPublished(t *testing.T) {
	f := newFakeRegistries(t)
	probe := f.probe()
	probe.RubyGemsKey = "revoked"
	found := blocking(ValidateRegistries(allTargets(), probe))
	if len(found) != 1 || found[0].Registry != "RubyGems" {
		t.Fatalf("blocking findings: %+v", found)
	}
	if !strings.Contains(found[0].String(), "GEM_HOST_API_KEY") || !strings.Contains(found[0].String(), "401") {
		t.Errorf("the finding should name the secret to fix: %s", found[0])
	}
	if len(f.publishes) != 0 {
		t.Errorf("registries saw publishes: %v", f.publishes)
	}
}

func TestEachRegistryRefusesItsOwnBadCredential(t *testing.T) {
	cases := map[string]func(*RegistryProbe){
		"npm":      func(p *RegistryProbe) { p.NPMToken = "expired" },
		"PyPI":     func(p *RegistryProbe) { p.PyPIPassword = "expired" },
		"RubyGems": func(p *RegistryProbe) { p.RubyGemsKey = "expired" },
	}
	for registry, spoil := range cases {
		f := newFakeRegistries(t)
		probe := f.probe()
		spoil(probe)
		found := blocking(ValidateRegistries(allTargets(), probe))
		if len(found) != 1 || found[0].Registry != registry {
			t.Errorf("%s: blocking findings %+v", registry, found)
		}
	}
}

func TestNuGetRefusingTheEmptyPushIsOnlyAWarning(t *testing.T) {
	f := newFakeRegistries(t)
	probe := f.probe()
	probe.NuGetKey = "expired"
	found := ValidateRegistries(allTargets(), probe)
	if len(blocking(found)) != 0 {
		t.Fatalf("a NuGet answer must not stop the release: %+v", blocking(found))
	}
	warned := false
	for _, finding := range found {
		warned = warned || (finding.Registry == "NuGet" && strings.Contains(finding.Message, "not confirmed"))
	}
	if !warned {
		t.Errorf("the NuGet answer should still be said: %+v", found)
	}
}

func TestAnUnreachableRegistryStopsTheRelease(t *testing.T) {
	f := newFakeRegistries(t)
	probe := f.probe()
	f.server.Close()
	found := blocking(ValidateRegistries(allTargets()[:1], probe))
	if len(found) != 1 || !strings.Contains(found[0].Message, "unreachable") {
		t.Errorf("findings: %+v", found)
	}
}

func TestAVersionNpmAlreadyHasIsAWarningThatTheJournalExplains(t *testing.T) {
	f := newFakeRegistries(t)
	f.npmVersions = []string{"1.2.0"}
	probe := f.probe()
	found := ValidateRegistries(allTargets()[:1], probe)
	if len(found) != 1 || found[0].Blocking || !strings.Contains(found[0].Message, "not in the release journal") {
		t.Fatalf("unexplained: %+v", found)
	}
	probe.Published = func(_, _, name string) bool { return name == "acme" }
	found = ValidateRegistries(allTargets()[:1], probe)
	if len(found) != 1 || found[0].Blocking || !strings.Contains(found[0].Message, "already published by this release") {
		t.Errorf("explained by the journal: %+v", found)
	}
}

func TestANpmNameSomeoneElseMaintainsBlocksOnlyWhenItIsNotScoped(t *testing.T) {
	f := newFakeRegistries(t)
	f.npmMaintainers = []string{"someone-else"}
	found := blocking(ValidateRegistries(allTargets()[:1], f.probe()))
	if len(found) != 1 || !strings.Contains(found[0].Message, "not a maintainer") {
		t.Errorf("findings: %+v", found)
	}
}

func appSession(t *testing.T, release *workspace.ProjectRelease) (*session.Session, Plan) {
	t.Helper()
	g, err := projectgraph.New([]*projectgraph.Project{{
		Project: protocol.Project{Name: "cli", Root: "apps/cli", Kind: protocol.ProjectApp, Version: "0.0.0"},
		Plugin:  "test", Release: release,
	}})
	if err != nil {
		t.Fatal(err)
	}
	s := &session.Session{Root: t.TempDir(), Git: repo(t), Graph: g}
	return s, Plan{Items: []Item{{Project: "cli", Plugin: "test", Next: "1.2.0", Tag: "cli@1.2.0"}}}
}

func TestTheReleaseIsRefusedBeforeItsFirstPublishWhenALaterRegistryRefusesTheKey(t *testing.T) {
	f := newFakeRegistries(t)
	probe := f.probe()
	probe.RubyGemsKey = "revoked"
	s, plan := appSession(t, &workspace.ProjectRelease{NPM: []string{"acme"}, PyPI: "acme", NuGet: "acme", RubyGems: "acme"})
	var out strings.Builder
	err := checkRegistries(s, plan, probe, &out)
	if err == nil || !strings.Contains(err.Error(), "nothing was published") || !strings.Contains(err.Error(), "RubyGems (cli)") {
		t.Fatalf("error %v", err)
	}
	if len(f.publishes) != 0 {
		t.Errorf("publishes: %v", f.publishes)
	}
}

func TestAPlanWithoutRegistriesIsNotChecked(t *testing.T) {
	s, plan := appSession(t, &workspace.ProjectRelease{})
	if err := checkRegistries(s, plan, &RegistryProbe{}, &strings.Builder{}); err != nil {
		t.Error(err)
	}
}

func TestAPushOnlyRubyGemsKeyIsAWarningNotABlock(t *testing.T) {
	f := newFakeRegistries(t)
	f.rubyPushOnly = "push-only"
	probe := f.probe()
	probe.RubyGemsKey = "push-only"
	found := ValidateRegistries(allTargets(), probe)
	if len(blocking(found)) != 0 {
		t.Fatalf("a valid push-only key blocked the release: %+v", found)
	}
	warned := false
	for _, x := range found {
		warned = warned || (x.Registry == "RubyGems" && strings.Contains(x.String(), "index_rubygems"))
	}
	if !warned {
		t.Errorf("the missing scope should be said: %+v", found)
	}
}
