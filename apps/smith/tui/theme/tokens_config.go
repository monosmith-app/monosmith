package theme

// Tokens are the semantic colours of one theme, as Lip Gloss colour specs
// (ANSI 256 numbers or #rrggbb). Components ask for a meaning (success, error,
// accent), never for a raw colour. An empty colour means "the terminal's own".
type Tokens struct {
	Name string
	// Accent marks selection, focus and the active tab.
	Accent string
	// Success is passed, cached and ready; Warning needs attention; Error is
	// failed or blocked; Info is running or informational.
	Success, Warning, Error, Info string
	// Muted is skipped, disabled and secondary text; Faint is rules and hints.
	Muted, Faint string
	// Text is the normal text colour.
	Text string
	// Highlight is the background of the hovered or selected row, and Bar the
	// background of the status bar.
	Highlight, Bar string
	// Border is the rule of an unfocused pane and Focus the rule of the focused one.
	Border, Focus string
	// Colour is false for the plain theme: attributes (bold, reverse) carry
	// the meaning instead.
	Colour bool
}

// Dark suits a dark terminal background.
var Dark = Tokens{
	Name: "dark", Accent: "212", Success: "114", Warning: "215", Error: "203", Info: "75",
	Muted: "245", Faint: "240", Text: "252", Highlight: "238", Bar: "236",
	Border: "240", Focus: "212", Colour: true,
}

// Light suits a light terminal background; the colours keep a contrast of 4.5:1.
var Light = Tokens{
	Name: "light", Accent: "125", Success: "28", Warning: "130", Error: "160", Info: "25",
	Muted: "241", Faint: "243", Text: "235", Highlight: "253", Bar: "254",
	Border: "243", Focus: "125", Colour: true,
}

// Plain has no colour at all (NO_COLOR): bold and reverse video mark the
// selection, and every state also has a symbol and a word.
var Plain = Tokens{Name: "plain"}
