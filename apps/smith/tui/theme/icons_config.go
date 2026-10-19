package theme

// Icons are the symbols of the shell; every state has one, so colour is never
// the only signal.
type Icons struct {
	Passed, Failed, Warning, Running, Queued, Cached, Skipped string
	Dot, Open, Closed, Pointer, Bar, BarEmpty                 string
	Spinner                                                   []string
}

// Unicode symbols for terminals that draw them.
var Unicode = Icons{
	Passed: "✔", Failed: "✖", Warning: "▲", Running: "●", Queued: "○", Cached: "⚡", Skipped: "↷",
	Dot: "●", Open: "▾", Closed: "▸", Pointer: "›", Bar: "█", BarEmpty: "░",
	Spinner: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
}

// ASCII are the fallbacks for terminals without Unicode.
var ASCII = Icons{
	Passed: "[v]", Failed: "[x]", Warning: "[!]", Running: "[~]", Queued: "[ ]", Cached: "[c]", Skipped: "[-]",
	Dot: "*", Open: "v", Closed: ">", Pointer: ">", Bar: "#", BarEmpty: "-",
	Spinner: []string{"|", "/", "-", "\\"},
}
