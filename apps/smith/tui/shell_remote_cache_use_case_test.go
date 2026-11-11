package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

// remoteLab is Home with a remote cache whose probe answers status (after
// hold closes, when hold is set) and whose count is pending.
func remoteLab(t *testing.T, w int, status remotesync.Status, pending sources.RemotePending, hold chan struct{}) *lab {
	t.Helper()
	l := newHomeLab(t, w, 40)
	sh := l.sh()
	sh.sources.RemoteSet = func() bool { return true }
	sh.sources.RemoteStatus = func() remotesync.Status {
		if hold != nil {
			<-hold
		}
		return status
	}
	sh.sources.RemotePending = func() sources.RemotePending { return pending }
	cmd := sh.checkRemote()
	l.m = sh
	l.drive(cmd)
	return l
}

func connected(n int, atLeast bool) (remotesync.Status, sources.RemotePending) {
	return remotesync.Status{State: remotesync.StateConnected}, sources.RemotePending{Pending: remotesync.Pending{Count: n, AtLeast: atLeast, Checked: n}}
}

func TestTheHeaderShowsTheRemoteCacheStateInBothLayouts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status remotesync.Status
		mark   string
	}{
		{"connected", remotesync.Status{State: remotesync.StateConnected}, "≈✓"},
		{"unreachable", remotesync.Status{State: remotesync.StateUnreachable, Detail: "no answer"}, "≈▲"},
		{"refused", remotesync.Status{State: remotesync.StateRefused, Detail: "HEAD ac/0: 403 Forbidden"}, "≈×"},
	} {
		for _, w := range []int{160, 110} {
			l := remoteLab(t, w, tc.status, sources.RemotePending{}, nil)
			first := l.lines()[0]
			if !strings.Contains(first, tc.mark) {
				t.Errorf("%s at %d columns: the first row lacks %q: %q", tc.name, w, tc.mark, first)
			}
			if got := pane.Width(l.sh().headerText()); got > inlineHeaderMax {
				t.Errorf("%s: the header takes %d columns, more than %d", tc.name, got, inlineHeaderMax)
			}
			if w == 160 && !l.sh().inlineHeader() {
				t.Errorf("%s: the remote mark pushed the header off the tab row", tc.name)
			}
		}
	}
}

func TestTheHeaderSaysCheckingUntilTheProbeEnds(t *testing.T) {
	hold := make(chan struct{})
	defer close(hold)
	s, p := connected(0, false)
	l := remoteLab(t, 160, s, p, hold)
	if first := l.lines()[0]; !strings.Contains(first, "≈ checking…") {
		t.Fatalf("the header must say checking while the probe runs: %q", first)
	}
	if !strings.Contains(strings.Join(l.lines(), "\n"), "Remote cache: checking…") {
		t.Fatalf("Home must say checking:\n%s", strings.Join(l.lines(), "\n"))
	}
}

func TestWithoutARemoteTheHeaderHasNoRemoteMark(t *testing.T) {
	l := newHomeLab(t, 160, 40)
	if first := l.lines()[0]; strings.Contains(first, "≈") {
		t.Fatalf("no remote, no mark: %q", first)
	}
	if strings.Contains(strings.Join(l.lines(), "\n"), "Remote cache") {
		t.Fatal("no remote, no line on Home")
	}
}

func TestHomeSaysHowManyLocalEntriesTheRemoteLacks(t *testing.T) {
	for _, tc := range []struct {
		n       int
		atLeast bool
		want    string
	}{
		{12, false, "Remote cache: connected · 12 local entries not uploaded"},
		{500, true, "Remote cache: connected · at least 500 local entries not uploaded"},
		{0, false, "Remote cache: connected · it holds every local entry"},
	} {
		s, p := connected(tc.n, tc.atLeast)
		page := strings.Join(remoteLab(t, 160, s, p, nil).lines(), "\n")
		if !strings.Contains(page, tc.want) {
			t.Errorf("Home lacks %q:\n%s", tc.want, page)
		}
		if tc.n > 0 && !strings.Contains(page, "Push") {
			t.Errorf("Home must offer the push:\n%s", page)
		}
	}
}

func TestHomeSaysTheRemoteRefusesTheCredentials(t *testing.T) {
	page := strings.Join(remoteLab(t, 160, remotesync.Status{State: remotesync.StateRefused, Detail: "HEAD ac/0: 403 Forbidden"}, sources.RemotePending{}, nil).lines(), "\n")
	for _, want := range []string{"Remote cache: refuses the credentials", "Set up again"} {
		if !strings.Contains(page, want) {
			t.Errorf("Home lacks %q:\n%s", want, page)
		}
	}
}

func TestHomeSaysWhenTheCountFailed(t *testing.T) {
	s, _ := connected(0, false)
	page := strings.Join(remoteLab(t, 160, s, sources.RemotePending{Err: errors.New("disk")}, nil).lines(), "\n")
	if !strings.Contains(page, "could not count") {
		t.Errorf("Home must say the count failed:\n%s", page)
	}
}
