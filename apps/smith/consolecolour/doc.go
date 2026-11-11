// Package consolecolour makes a Windows console render the ANSI colour codes
// smith and the tools it runs print. Windows Terminal, VS Code and mintty
// render them anyway; the classic console (conhost) does so only once a
// program turns on virtual terminal processing. On other systems it does
// nothing.
package consolecolour
