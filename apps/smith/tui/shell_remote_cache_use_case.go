package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dashboard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// remoteCache is what the shell knows of the remote cache. Both questions run
// in the background, as doctor does: the header and Home say "checking" until
// they end.
type remoteCache struct {
	// set is true when the workspace has a remote cache; the rest is empty
	// without one.
	set bool
	// probing is true while the probe runs; status is its answer.
	probing bool
	status  *remotesync.Status
	// counting is true while the local entries the remote lacks are counted;
	// pending is the count.
	counting bool
	pending  *sources.RemotePending
}

type (
	// remoteStatusMsg is the answer of the probe.
	remoteStatusMsg struct{ status remotesync.Status }
	// remotePendingMsg is the count of the local entries the remote lacks.
	remotePendingMsg struct{ pending sources.RemotePending }
)

// checkRemote starts the probe of the remote cache, when the workspace has
// one; it forgets what an earlier probe found.
func (s *Shell) checkRemote() tea.Cmd {
	set := s.sources.RemoteSet != nil && s.sources.RemoteSet()
	if !set || s.sources.RemoteStatus == nil {
		s.remote = remoteCache{}
		return nil
	}
	if s.remote.probing {
		return nil
	}
	s.remote = remoteCache{set: true, probing: true}
	probe := s.sources.RemoteStatus
	return func() tea.Msg { return remoteStatusMsg{status: probe()} }
}

// remoteChecked keeps the probe's answer and, when the remote is connected,
// starts the count.
func (s *Shell) remoteChecked(msg remoteStatusMsg) tea.Cmd {
	s.remote.probing, s.remote.status = false, &msg.status
	if msg.status.State == remotesync.StateNone {
		s.remote = remoteCache{}
		return nil
	}
	return s.countRemote()
}

// countRemote starts the count of the local entries a connected remote lacks.
func (s *Shell) countRemote() tea.Cmd {
	count := s.sources.RemotePending
	if count == nil || s.remote.counting || s.remote.status == nil || s.remote.status.State != remotesync.StateConnected {
		return nil
	}
	s.remote.counting = true
	return func() tea.Msg { return remotePendingMsg{pending: count()} }
}

// remoteCounted keeps the count.
func (s *Shell) remoteCounted(msg remotePendingMsg) {
	s.remote.counting, s.remote.pending = false, &msg.pending
}

// remoteFacts is the remote cache as Home tells it; nil without a remote.
func (s Shell) remoteFacts() *dashboard.Remote {
	if !s.remote.set {
		return nil
	}
	r := &dashboard.Remote{State: dashboard.RemoteChecking}
	st := s.remote.status
	if st == nil {
		return r
	}
	r.Detail = st.Detail
	switch st.State {
	case remotesync.StateConnected:
		r.State = dashboard.RemoteConnected
	case remotesync.StateUnreachable:
		r.State = dashboard.RemoteUnreachable
	case remotesync.StateRefused:
		r.State = dashboard.RemoteRefused
	}
	if p := s.remote.pending; p != nil && !s.remote.counting {
		r.Counted, r.Pending, r.AtLeast = true, p.Pending.Count, p.Pending.AtLeast
		if p.Err != nil {
			r.CountErr = p.Err.Error()
		}
	}
	return r
}

// remoteSummary is the remote cache in the header: the cache mark and the
// state's mark (connected, does not answer, refuses the credentials), or
// "checking…" until the probe ends; "" without a remote.
func (s Shell) remoteSummary() string {
	if !s.remote.set {
		return ""
	}
	icons := s.th.Icons
	st := s.remote.status
	if st == nil {
		return s.th.Muted(words.Say("tui.header.remote_checking", "icon", icons.Cached))
	}
	switch st.State {
	case remotesync.StateConnected:
		return s.th.Success(icons.Cached + icons.Passed)
	case remotesync.StateUnreachable:
		return s.th.Warning(icons.Cached + icons.Warning)
	case remotesync.StateRefused:
		return s.th.Failure(icons.Cached + icons.Failed)
	}
	return ""
}
