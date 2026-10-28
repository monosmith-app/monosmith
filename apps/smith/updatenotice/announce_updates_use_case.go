package updatenotice

import (
	"context"
	"fmt"
	"time"
)

// checkEvery is how often the registry is asked, and how often a newer version
// is pointed out.
const checkEvery = 24 * time.Hour

// Options say what runs and where it keeps its memory.
type Options struct {
	// Home is the smith home; "" remembers nothing and so says nothing.
	Home string
	// Version is the smith that is running.
	Version string
	// Exe is the path of the running binary, to pick the update command.
	Exe, GOOS string
	// Off is true in CI: nothing is read, said or written.
	Off bool
	// NoCheck is true when the person turned the registry check off
	// (SMITH_NO_UPDATE_CHECK=1). A version change is still announced.
	NoCheck bool
	// Now is the clock; time.Now when nil.
	Now func() time.Time
	// Fetch asks the registry for the newest version; FetchLatest with a short
	// timeout when nil.
	Fetch func(ctx context.Context) (string, error)
	// Wait is how long Lines waits for a check still under way; 0 does not wait.
	Wait time.Duration
}

// Pending is an announcement being worked out: the upgrade line is known at
// once, the registry answer arrives later or never.
type Pending struct {
	o       Options
	upgrade string
	done    chan struct{}
}

// Start notes which smith runs now and, when a day has passed since the last
// look, asks the registry in the background. It never blocks.
func Start(o Options) *Pending {
	if o.Now == nil {
		o.Now = time.Now
	}
	p := &Pending{o: o}
	if _, ok := parse(o.Version); !ok || o.Home == "" {
		p.o.Off = true
	}
	if p.o.Off {
		return p
	}
	m := load(o.Home)
	if Newer(o.Version, m.LastVersion) {
		p.upgrade = fmt.Sprintf("smith updated %s → %s. What's new: %s/tag/smith@%s", m.LastVersion, o.Version, ReleasesURL, o.Version)
	}
	if m.LastVersion != o.Version {
		update(o.Home, func(m *memory) { m.LastVersion = o.Version })
	}
	if o.NoCheck || o.Now().Sub(m.CheckedAt) < checkEvery {
		return p
	}
	fetch := o.Fetch
	if fetch == nil {
		fetch = func(ctx context.Context) (string, error) { return FetchLatest(ctx, 2*time.Second) }
	}
	// Marked before the question, so a slow or missing network is tried once a
	// day and not on every command.
	update(o.Home, func(m *memory) { m.CheckedAt = o.Now() })
	p.done = make(chan struct{})
	go func() {
		defer close(p.done)
		if latest, err := fetch(context.Background()); err == nil {
			update(o.Home, func(m *memory) { m.Latest = latest })
		}
	}()
	return p
}

// Lines are the quiet lines to show, if any: that smith was updated (once), and
// that a newer one exists (at most once a day).
func (p *Pending) Lines() []string {
	var lines []string
	if p.upgrade != "" {
		lines = append(lines, p.upgrade)
	}
	if p.done != nil && p.o.Wait > 0 {
		select {
		case <-p.done:
		case <-time.After(p.o.Wait):
		}
	}
	if p.o.Off || p.o.NoCheck || p.o.Home == "" {
		return lines
	}
	m := load(p.o.Home)
	if !Newer(m.Latest, p.o.Version) || p.o.Now().Sub(m.NoticeAt) < checkEvery {
		return lines
	}
	update(p.o.Home, func(m *memory) { m.NoticeAt = p.o.Now() })
	return append(lines, fmt.Sprintf("smith %s is out, you have %s: update with %s (hide this: SMITH_NO_UPDATE_CHECK=1)",
		m.Latest, p.o.Version, InstallCommand(p.o.Exe, p.o.GOOS)))
}
