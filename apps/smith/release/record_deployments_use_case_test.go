package release

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
)

func deploymentServer(t *testing.T, status int) (*githubrelease.Client, *[]map[string]any) {
	t.Helper()
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		bodies = append(bodies, body)
		if status >= 300 {
			http.Error(w, `{"message":"Resource not accessible by integration"}`, status)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": 7}`))
	}))
	t.Cleanup(server.Close)
	return &githubrelease.Client{Token: "tok", API: server.URL, HTTP: server.Client()}, &bodies
}

func TestEachRegistryOfTheRealSampleBecomesOneDeployment(t *testing.T) {
	client, bodies := deploymentServer(t, http.StatusCreated)
	var out bytes.Buffer
	recordWith(client, githubrelease.Repository{Owner: "monosmith-app", Name: "monosmith"}, sampleReport(), &out)
	if out.Len() != 0 {
		t.Errorf("a recorded release says nothing: %q", out.String())
	}
	var got []string
	for _, b := range *bodies {
		if env, ok := b["environment"].(string); ok {
			got = append(got, env+"|"+b["ref"].(string)+"|"+b["description"].(string))
		}
		if url, ok := b["environment_url"].(string); ok {
			got = append(got, url)
		}
	}
	want := []string{
		"npm|smith@1.129.0|@monosmith/cli 1.129.0", "https://www.npmjs.com/package/@monosmith/cli/v/1.129.0",
		"PyPI|smith@1.129.0|monosmith 1.129.0", "https://pypi.org/project/monosmith/1.129.0",
		"NuGet|smith@1.129.0|monosmith 1.129.0", "https://www.nuget.org/packages/monosmith/1.129.0",
		"RubyGems|smith@1.129.0|monosmith 1.129.0", "https://rubygems.org/gems/monosmith/versions/1.129.0",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s", strings.Join(got, "\n"))
	}
}

func TestAForbiddenDeploymentOnlyWarnsOnceWithTheRemedy(t *testing.T) {
	client, bodies := deploymentServer(t, http.StatusForbidden)
	var out bytes.Buffer
	recordWith(client, githubrelease.Repository{Owner: "a", Name: "b"}, sampleReport(), &out)
	if len(*bodies) != 1 {
		t.Errorf("the first refusal stops the rest: %d requests", len(*bodies))
	}
	line := out.String()
	if strings.Count(line, "\n") != 1 || !strings.HasPrefix(line, "warning: could not record the release as GitHub deployments: npm @monosmith/cli 1.129.0") ||
		!strings.Contains(line, "deployments: write") || !strings.Contains(line, "github_deployments = false") {
		t.Errorf("%q", line)
	}
}

func TestNothingIsRecordedForWhatWasNotPublishedNow(t *testing.T) {
	failed := sampleReport()
	failed.Results[0].Status = "failed"
	unpushed := sampleReport()
	unpushed.Results[0].Pushed = false
	already := Report{Results: []ReportResult{{Item: Item{Project: "k", Next: "1.0.0", Tag: "k@1.0.0"}, Status: "released", Pushed: true, ExistingPackages: []string{"k@1.0.0"}}}}
	for name, r := range map[string]Report{"failed": failed, "tag not pushed": unpushed, "registry already had it": already} {
		if len(deployable(r)) != 0 {
			t.Errorf("%s: %+v", name, deployable(r))
		}
	}
}

func TestLocalRunsAndOtherCIsRecordNothing(t *testing.T) {
	var out bytes.Buffer
	for _, env := range []string{"", "false"} {
		t.Setenv("GITHUB_ACTIONS", env)
		// A nil session proves nothing is read: the check stops at the environment.
		recordDeployments(nil, sampleReport(), &out)
	}
	t.Setenv("GITHUB_ACTIONS", "true")
	recordDeployments(nil, Report{DryRun: true, Results: sampleReport().Results}, &out)
	if out.Len() != 0 {
		t.Errorf("%q", out.String())
	}
}
