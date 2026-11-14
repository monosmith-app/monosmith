package updatenotice

// Update is what smith knows about a newer smith: the running version, the
// newest one the registry reported on the person's channel, and the command
// that installs it the way smith was installed.
type Update struct {
	// Current is the smith that is running.
	Current string
	// Latest is the newest version on Channel; "" when the registry has not been asked yet.
	Latest string
	// Channel is the dist-tag followed: "latest" or "next".
	Channel string
	// Command is the command that updates smith the way it was installed, or
	// "see <releases page>" when smith cannot tell how it was installed.
	Command string
	// Available is true when Latest is newer than Current.
	Available bool
	// Alerted is true when the interface already offered Latest in its alert.
	Alerted bool
	// Muted is true when the person asked not to be reminded: the interface
	// keeps the update out of the status bar. The check itself goes on.
	Muted bool
}

// Runnable is true when Command can be run: smith knows how it was installed.
func (u Update) Runnable() bool { return u.Command != "" && !isReleasePage(u.Command) }
