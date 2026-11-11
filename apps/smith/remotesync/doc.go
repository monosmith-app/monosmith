// Package remotesync shows and changes what the remote task cache holds, from
// the outside of a run: whether the remote answers and accepts the
// credentials, how many local entries it lacks, and the push that sends them.
// It never opens a smith session, so the TUI can call it in the background and
// a test can run it against an httptest server.
package remotesync
