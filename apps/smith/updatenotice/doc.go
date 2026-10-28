// Package updatenotice tells a person when smith changed under them: once
// after the installed smith is newer than the one they last ran, and at most
// once a day when the registry has a newer one than the one running. Both are
// one quiet line. The check is cached in the smith home, has a short timeout,
// never blocks a command, says nothing offline, and is off in CI and with
// SMITH_NO_UPDATE_CHECK=1.
package updatenotice
