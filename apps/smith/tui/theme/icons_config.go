package theme

import "github.com/monosmith-app/monosmith/apps/smith/marks"

// Icons are the symbols of the shell; every state has one, so colour is never
// the only signal.
type Icons struct {
	Passed, Failed, Warning, Running, Queued, Cached, Skipped, Cancelled string
	Dot, Open, Closed, Pointer, Bar, BarEmpty                            string
	// Track and Thumb draw a scrollbar: the dashed line, and the solid, lighter
	// part of it that shows. The thumb is never the heavy divider glyph.
	Track, Thumb string
	Spinner      []string
}

// Unicode symbols for terminals that draw them.
var Unicode = Icons{
	Passed: marks.Passed, Failed: marks.Failed, Warning: marks.Warning, Running: "●", Queued: "○", Cached: marks.Cached, Skipped: marks.Skipped, Cancelled: marks.Cancelled,
	Dot: "●", Open: "▾", Closed: "▸", Pointer: "›", Bar: "█", BarEmpty: "░", Track: "┆", Thumb: "│",
	Spinner: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
}

// ASCII are the fallbacks for terminals without Unicode.
var ASCII = Icons{
	Passed: "[v]", Failed: "[x]", Warning: "[!]", Running: "[~]", Queued: "[ ]", Cached: "[c]", Skipped: "[-]", Cancelled: "[/]",
	Dot: "*", Open: "v", Closed: ">", Pointer: ">", Bar: "#", BarEmpty: "-", Track: "|", Thumb: "#",
	Spinner: []string{"|", "/", "-", "\\"},
}
