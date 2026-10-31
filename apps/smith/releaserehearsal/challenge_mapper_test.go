package releaserehearsal

import "testing"

func TestParseChallenge(t *testing.T) {
	c := ParseChallenge(`Basic realm="Azure DevOps Services", Bearer`)
	if c.Scheme != "Basic" || c.Realm != "Azure DevOps Services" {
		t.Fatalf("got %+v", c)
	}
	if got := ParseChallenge("Bearer").Expects(); got != "Bearer (a token)" {
		t.Fatalf("got %q", got)
	}
	if got := ParseChallenge("").Expects(); got != "no credential to read" {
		t.Fatalf("got %q", got)
	}
}
