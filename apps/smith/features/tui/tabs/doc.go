// Package tabs draws the shell's row of views as a strip of rounded tabs whose
// active tab is open into the bordered pane below it (three rows), or, on a
// short terminal, as a single row of names. When the names do not fit they are
// shortened first, then the strip scrolls with ‹ and › so the active tab is
// always in view. It reports where each tab sits so a click can be mapped to
// it, and draws nothing but text: the shell owns the state.
package tabs
