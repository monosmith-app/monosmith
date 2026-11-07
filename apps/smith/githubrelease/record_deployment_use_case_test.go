package githubrelease

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecordDeploymentCreatesTheDeploymentThenMarksItSuccessful(t *testing.T) {
	var calls []string
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		calls, bodies = append(calls, r.Method+" "+r.URL.Path), append(bodies, body)
		if strings.HasSuffix(r.URL.Path, "/deployments") {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": 42}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	c := &Client{Token: "tok", API: server.URL, HTTP: server.Client()}
	err := c.RecordDeployment(Repository{Owner: "acme", Name: "tool"}, Deployment{
		Environment: "npm", Ref: "tool@1.2.0", Description: "@acme/tool 1.2.0", URL: "https://www.npmjs.com/package/@acme/tool/v/1.2.0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0] != "POST /repos/acme/tool/deployments" || calls[1] != "POST /repos/acme/tool/deployments/42/statuses" {
		t.Fatalf("calls %v", calls)
	}
	if bodies[0]["environment"] != "npm" || bodies[0]["ref"] != "tool@1.2.0" || bodies[0]["description"] != "@acme/tool 1.2.0" || bodies[0]["auto_merge"] != false {
		t.Errorf("deployment body %v", bodies[0])
	}
	if contexts, ok := bodies[0]["required_contexts"].([]any); !ok || len(contexts) != 0 {
		t.Errorf("a published package must not wait for status checks: %v", bodies[0])
	}
	if bodies[1]["state"] != "success" || bodies[1]["environment_url"] != "https://www.npmjs.com/package/@acme/tool/v/1.2.0" {
		t.Errorf("status body %v", bodies[1])
	}
}

func TestRecordDeploymentReportsARefusalAndSendsNoStatus(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		http.Error(w, `{"message":"Resource not accessible by integration"}`, http.StatusForbidden)
	}))
	defer server.Close()
	c := &Client{Token: "tok", API: server.URL, HTTP: server.Client()}
	err := c.RecordDeployment(Repository{Owner: "acme", Name: "tool"}, Deployment{Environment: "npm", Ref: "t"})
	if err == nil || !strings.Contains(err.Error(), "403") || requests != 1 {
		t.Fatalf("err %v after %d requests", err, requests)
	}
}
