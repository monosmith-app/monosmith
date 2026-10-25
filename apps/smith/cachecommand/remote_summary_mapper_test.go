package cachecommand

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

func TestARemoteIsDescribedByItsSchemeAndHostOnly(t *testing.T) {
	for _, tc := range []struct {
		url, scheme, host string
	}{
		{"https://cache.example.com/team/main?sig=SECRET", "https", "cache.example.com"},
		{"https://user:hunter2@cache.example.com:8443/x", "https", "cache.example.com:8443"},
		{"s3://acme-smith-cache/main", "s3", "acme-smith-cache"},
		{"azblob://acmestore/cache/main", "azblob", "acmestore"},
	} {
		got := summarizeRemote(session.RemotePlan{URL: tc.url, Active: true})
		if got.Scheme != tc.scheme || got.Host != tc.host || !got.Configured || !got.Active {
			t.Errorf("%s: %+v", tc.url, got)
		}
		raw, _ := json.Marshal(got)
		for _, secret := range []string{"SECRET", "hunter2", "/team", "main"} {
			if strings.Contains(string(raw), secret) {
				t.Errorf("%s: the summary %s leaks %q", tc.url, raw, secret)
			}
		}
	}
}

func TestNoRemoteIsNotConfigured(t *testing.T) {
	if got := summarizeRemote(session.RemotePlan{}); got.Configured || got.Scheme != "" || got.Host != "" {
		t.Fatalf("%+v", got)
	}
}
