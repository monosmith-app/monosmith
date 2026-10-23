package release

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestTheRubyGemsKeyIsSignedInByTheRunOnlyWithTrustedPublishing(t *testing.T) {
	cfg := workspace.Config{}
	secret := &session.Session{Config: cfg}
	cfg.Release.RubyGemsTrustedPublishing = true
	trusted := &session.Session{Config: cfg}
	cases := []struct {
		s    *session.Session
		env  string
		want bool
	}{
		{secret, "GEM_HOST_API_KEY", false},
		{trusted, "GEM_HOST_API_KEY", true},
		{trusted, "NODE_AUTH_TOKEN", false},
	}
	for _, c := range cases {
		if got := signedInByTheRun(c.s, c.env); got != c.want {
			t.Errorf("signedInByTheRun(trusted=%v, %s) = %v, want %v", c.s.Config.Release.RubyGemsTrustedPublishing, c.env, got, c.want)
		}
	}
}
