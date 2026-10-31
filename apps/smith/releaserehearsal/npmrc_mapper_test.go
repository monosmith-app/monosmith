package releaserehearsal

import "testing"

func TestNpmrcRegistry(t *testing.T) {
	rc := "registry=https://registry.npmjs.org/\n@acme:registry=https://pkgs.dev.azure.com/acme/_packaging/feed/npm/registry/\n"
	if got := NpmrcRegistry(rc, "@acme/core"); got != "https://pkgs.dev.azure.com/acme/_packaging/feed/npm/registry/" {
		t.Fatalf("scoped: %q", got)
	}
	if got := NpmrcRegistry(rc, "core"); got != "https://registry.npmjs.org/" {
		t.Fatalf("unscoped: %q", got)
	}
	if got := NpmrcRegistry("", "core"); got != "" {
		t.Fatalf("none: %q", got)
	}
}

func TestNpmrcToken(t *testing.T) {
	rc := "//pkgs.dev.azure.com/acme/_packaging/feed/npm/registry/:_authToken=${FEED_TOKEN}\n//registry.npmjs.org/:_authToken=literal\n"
	env := envOf(map[string]string{"FEED_TOKEN": "abc"})
	tok, v := NpmrcToken(rc, "https://pkgs.dev.azure.com/acme/_packaging/feed/npm/registry/", env)
	if tok != "abc" || v != "FEED_TOKEN" {
		t.Fatalf("got %q %q", tok, v)
	}
	if tok, v = NpmrcToken(rc, "https://registry.npmjs.org", env); tok != "literal" || v != "" {
		t.Fatalf("got %q %q", tok, v)
	}
	if tok, _ = NpmrcToken(rc, "https://other.example", env); tok != "" {
		t.Fatalf("an unrelated host has no token: %q", tok)
	}
}
