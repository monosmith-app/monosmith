// Package runner starts a smith command as a child process (the smith binary
// itself, so the shell never does what a command line could not), streams its
// output as lines, and stops it on request. It has no terminal or Bubble Tea
// in it: the shell turns Next into a command.
package runner
