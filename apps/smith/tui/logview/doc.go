// Package logview is the scrolling log of a run: it follows the end while new
// lines arrive, pauses when the user scrolls up, and colours the lines that
// say how a task went. The lines themselves belong to the run; this package
// keeps only the scroll position.
package logview
