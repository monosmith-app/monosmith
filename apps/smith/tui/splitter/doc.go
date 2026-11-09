// Package splitter is the one way panels are resized in the TUI. A screen lists
// its dividers as Spots (a column between two panels side by side, a row between
// two stacked ones). State then does the rest: it selects a divider (Select,
// Cycle), moves it with the keys (Move), grabs and drags it with the pointer
// (Press, Drag, Release), keeps each size between its limits, and Paint draws
// the selected or held divider. A screen never moves a divider by hand.
package splitter
