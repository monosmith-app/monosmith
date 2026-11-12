package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/cachesetup"
	"github.com/monosmith-app/monosmith/apps/smith/tui/cachewizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// cacheMsg is news from the engine of the remote cache setup.
type cacheMsg struct {
	run *cachewizard.Bridge
	ev  cachewizard.Event
}

// waitCache blocks, off the Update loop, until the engine has news.
func waitCache(b *cachewizard.Bridge) tea.Cmd {
	return func() tea.Msg {
		select {
		case ev := <-b.Events:
			return cacheMsg{run: b, ev: ev}
		case <-b.Done():
			return nil
		}
	}
}

// openCacheSetup opens the remote cache setup at once, on a spinner, and runs
// the same engine as `smith cache setup` behind it.
func (s *Shell) openCacheSetup() tea.Cmd {
	setup := s.sources.CacheSetup
	if setup == nil {
		s.say(words.Say("tui.cache.not_here"))
		return nil
	}
	if s.cache != nil {
		return nil
	}
	secrets := cachewizard.SystemSecrets(s.clip)
	if s.cacheSecrets != nil {
		secrets = s.cacheSecrets()
	}
	b := cachewizard.NewBridge(secrets)
	m := cachewizard.New(s.now()).Place(s.w, s.h)
	s.cache, s.cacheRun = &m, b
	b.Start(func(d cachesetup.Dialogue) (string, error) {
		facts, ports, err := setup()
		if err != nil {
			return "", err
		}
		ports.Dialogue = d
		return cachesetup.Run(facts, ports)
	})
	return waitCache(b)
}

// cacheKey gives the setup a key.
func (s *Shell) cacheKey(msg tea.KeyMsg) tea.Cmd {
	m, out := s.cache.Key(msg)
	s.cache = &m
	return s.afterCache(out)
}

// cacheClick gives the setup a click.
func (s *Shell) cacheClick(x, y int, double bool) tea.Cmd {
	m, out := s.cache.Click(s.th, x, y, double)
	s.cache = &m
	return s.afterCache(out)
}

// afterCache hands the engine what the screen decided.
func (s *Shell) afterCache(out cachewizard.Outcome) tea.Cmd {
	switch {
	case out.Cancel:
		s.cacheRun.Cancel()
		s.cache, s.cacheRun = nil, nil
		s.notify(words.Say("tui.cache.closed"), toast.Info)
	case out.Reply != nil:
		s.cacheRun.Reply(*out.Reply)
	case out.Press != 0:
		s.cacheRun.Press(out.Press)
	}
	return nil
}

// cacheEvent shows the engine's news, and closes the setup when it ends.
func (s *Shell) cacheEvent(msg cacheMsg) tea.Cmd {
	if msg.run != s.cacheRun || s.cache == nil {
		return nil
	}
	if end, ok := msg.ev.(cachewizard.EndEvent); ok {
		s.cache, s.cacheRun = nil, nil
		switch {
		case end.Err != nil:
			s.notify(words.Say("tui.cache.stopped", "reason", end.Err.Error()), toast.Error)
		case end.Remote != "":
			s.notify(words.Say("tui.cache.finished.for", "remote", end.Remote), toast.Success)
		default:
			s.notify(words.Say("tui.cache.finished"), toast.Success)
		}
		var reload tea.Cmd
		s.settings, reload = s.settings.Reload()
		// The remote may be new or changed: ask it again.
		return tea.Batch(reload, s.checkRemote())
	}
	m := s.cache.Handle(msg.ev, s.now())
	s.cache = &m
	return waitCache(msg.run)
}
