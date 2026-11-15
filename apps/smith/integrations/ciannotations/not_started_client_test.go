package ciannotations

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fakeGitHub(t *testing.T, annotation string) {
	t.Helper()
	routes := map[string]string{
		"/repos/acme/app":                          `{"default_branch":"main"}`,
		"/repos/acme/app/actions/runs":             `{"workflow_runs":[{"head_sha":"abc"}]}`,
		"/repos/acme/app/commits/abc/check-runs":   `{"check_runs":[{"id":1,"conclusion":"success"},{"id":7,"conclusion":"failure"}]}`,
		"/repos/acme/app/check-runs/7/annotations": `[{"message":"` + annotation + `"}]`,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		body, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	old := APIURL
	APIURL = srv.URL
	t.Cleanup(func() { APIURL = old; srv.Close() })
}

func TestTheNotStartedAnnotationIsReturned(t *testing.T) {
	msg := "The job was not started because recent account payments have failed or your spending limit needs to be increased."
	fakeGitHub(t, msg)
	got, err := NotStarted(nil, "acme", "app", "tok")
	if err != nil || got != msg {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestAnotherAnnotationIsNotReported(t *testing.T) {
	fakeGitHub(t, "Process completed with exit code 1.")
	got, err := NotStarted(nil, "acme", "app", "tok")
	if err != nil || got != "" {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestABadTokenIsAnError(t *testing.T) {
	fakeGitHub(t, "x")
	if _, err := NotStarted(nil, "acme", "app", "wrong"); err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("err = %v", err)
	}
}
