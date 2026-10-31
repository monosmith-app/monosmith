package pluginmarket

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func runSearch(t *testing.T, location string, configured string, args ...string) (string, error) {
	t.Helper()
	t.Setenv(LocationEnv, location)
	cmd := SearchCommand(http.DefaultClient, func() (string, error) { return configured, nil })
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func indexServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(sampleIndexJSON)) }))
	t.Cleanup(srv.Close)
	return srv
}

func TestSearchCommandListsTheNewestVersionOfEachMatch(t *testing.T) {
	srv := indexServer(t)
	out, err := runSearch(t, srv.URL+"/index.json", "", "scaffold")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"PLUGIN", "widget", "1.10.0", "scaffold", "Scaffolds widget projects", "smith plugin install"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "deployer") {
		t.Errorf("deployer does not match:\n%s", out)
	}
}

func TestSearchCommandReadsTheConfiguredIndexWhenTheEnvironmentIsEmpty(t *testing.T) {
	srv := indexServer(t)
	out, err := runSearch(t, "", srv.URL+"/index.json")
	if err != nil || !strings.Contains(out, "deployer") || !strings.Contains(out, "widget") {
		t.Fatalf("%v\n%s", err, out)
	}
}

func TestSearchCommandSaysWhenNothingMatchesAndPrintsJSON(t *testing.T) {
	srv := indexServer(t)
	out, err := runSearch(t, srv.URL, "", "zzz")
	if err != nil || !strings.Contains(out, "No plugin matches") {
		t.Fatalf("%v\n%s", err, out)
	}
	out, err = runSearch(t, srv.URL, "", "--json", "moon")
	if err != nil {
		t.Fatal(err)
	}
	var got []Plugin
	if err := json.Unmarshal([]byte(out), &got); err != nil || len(got) != 1 || got[0].Description != "Deploys to the moon" {
		t.Fatalf("%v %+v\n%s", err, got, out)
	}
	out, _ = runSearch(t, srv.URL, "", "--json", "zzz")
	if strings.TrimSpace(out) != "[]" {
		t.Fatalf("empty result must be [], got %q", out)
	}
}
