// Package button draws every clickable item of the TUI the same way: a label
// and its key in parentheses on a background that has three states, at rest,
// under the pointer and active (pressed, or the selected tab). Screens call
// Render for the look and Width for their click spans, and never draw a
// clickable item by hand.
package button
