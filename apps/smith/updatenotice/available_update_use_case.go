package updatenotice

import (
	"context"
	"time"
)

// Available reads what the last registry check found, without a request and
// without waiting: the interface asks it on every start and on every draw of
// Settings. It says nothing when the check is off, in CI, or read for another
// channel than the one smith runs on.
//
// Parameters: o are the options of the running smith (ForThisProcess).
// Returns: the update, and true when a newer smith is known.
func Available(o Options) (Update, bool) {
	u := Update{Current: o.Version, Channel: ChannelOf(o.Version), Command: UpdateCommand(o.Exe, o.GOOS, ChannelOf(o.Version))}
	if o.Off || o.NoCheck || o.Home == "" {
		return u, false
	}
	m := load(o.Home)
	u.Muted = m.Muted
	if m.Channel != "" && m.Channel != u.Channel {
		return u, false
	}
	u.Latest = m.Latest
	u.Available = Newer(m.Latest, o.Version)
	return u, u.Available
}

// Check asks the registry now, whatever the daily schedule and the person's
// update_check setting say: the person asked for it (smith update, the
// Settings section). The answer is kept, so the interface shows it too.
//
// Parameters: o are the options of the running smith; timeout bounds the request.
// Returns: the update (Available says whether a newer smith exists); an error
// when the registry does not answer.
func Check(ctx context.Context, o Options, timeout time.Duration) (Update, error) {
	channel := ChannelOf(o.Version)
	fetch := o.Fetch
	if fetch == nil {
		fetch = func(ctx context.Context) (string, error) { return FetchLatest(ctx, timeout, channel) }
	}
	latest, err := fetch(ctx)
	if err != nil {
		return Update{}, err
	}
	now := time.Now
	if o.Now != nil {
		now = o.Now
	}
	update(o.Home, func(m *memory) { m.Latest, m.Channel, m.CheckedAt = latest, channel, now() })
	u := Update{Current: o.Version, Latest: latest, Channel: channel, Command: UpdateCommand(o.Exe, o.GOOS, channel), Available: Newer(latest, o.Version)}
	u.Muted = load(o.Home).Muted
	return u, nil
}
