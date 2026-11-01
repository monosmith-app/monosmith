package adoptflow

import (
	"strings"
	"testing"
)

// A step runs without a terminal, so a repository with a root project always
// passes --move: the folder, or none. No session is opened.
func TestAdoptFixPassesTheAnswerAboutTheRootProject(t *testing.T) {
	s := &Session{}
	cases := []struct {
		commit bool
		move   string
		want   string
		desc   string
	}{
		{false, "apps/web", "adopt --init --move apps/web", "Move the project into apps/web"},
		{true, "apps/web", "adopt --init --move apps/web --commit", "the move is a commit of its own"},
		{false, "none", "adopt --init --move none", "leaving the project at the root"},
		{false, "", "adopt --init", "Make this repository a smith workspace"},
	}
	for _, c := range cases {
		fix := s.AdoptFix(c.commit, c.move)
		argv := strings.Join(fix.Steps[0].Argv[1:], " ")
		if argv != c.want || !strings.Contains(fix.Description, c.desc) {
			t.Errorf("AdoptFix(%v, %q) = %q / %q; want %q containing %q", c.commit, c.move, argv, fix.Description, c.want, c.desc)
		}
	}
}
