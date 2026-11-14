package updatenotice

// SetMuted remembers whether the person wants the update reminder kept out of
// the interface. It never stops the version check: the Settings section still
// shows the newest version, and smith update still works.
//
// Parameters: home is the smith home; muted is the person's choice.
func SetMuted(home string, muted bool) {
	update(home, func(m *memory) { m.Muted = muted })
}
