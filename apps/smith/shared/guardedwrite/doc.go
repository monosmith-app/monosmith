// Package guardedwrite writes a file the person named with -o without
// replacing another file silently: an existing file with other content is
// replaced only when the person passed the flag that says so, and the diff
// shows what is lost.
package guardedwrite
