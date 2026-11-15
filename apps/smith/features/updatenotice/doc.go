// Package updatenotice tells a person when smith changed under them: once
// after the installed smith is newer than the one they last ran, and at most
// once a day when the npm registry has a newer one on the channel they
// installed (stable follows the latest tag, a prerelease follows next). Both
// are one quiet line. The check is one request for the package's dist-tags,
// cached 24 hours in the smith home, with a one second timeout. It never blocks
// a command, says nothing offline, is off in CI, and is the person's choice:
// [workspace] update_check, SMITH_UPDATE_CHECK=on|off or SMITH_NO_UPDATE_CHECK=1.
// Until they choose, the first run announces the request and nothing is asked
// for a day.
package updatenotice
