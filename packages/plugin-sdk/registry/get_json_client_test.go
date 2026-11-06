package registry

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetJSONDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"version":"1.2.3"}`))
	}))
	defer srv.Close()
	var got struct{ Version string }
	if err := GetJSON(srv.URL, &got); err != nil || got.Version != "1.2.3" {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

func TestGetJSONNamesTheStatus(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	err := GetJSON(srv.URL, &struct{}{})
	if err == nil || !strings.Contains(err.Error(), "404") || !strings.Contains(err.Error(), srv.URL) {
		t.Fatalf("error = %v", err)
	}
}

func TestGetJSONRejectsBadBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	if err := GetJSON(srv.URL, &struct{}{}); err == nil {
		t.Fatal("want a decode error")
	}
}

func TestGetJSONUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	if err := GetJSON(url, &struct{}{}); err == nil {
		t.Fatal("want a connection error")
	}
}
