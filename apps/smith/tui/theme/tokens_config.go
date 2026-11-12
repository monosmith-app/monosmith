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
	// Highlight is the background of the selected row, RowHover the background
	// of a row under the pointer (between the terminal's own background and
	// Highlight), and Bar the background of the status bar.
	Highlight, RowHover, Bar string
	// Border is the rule of an unfocused pane and Focus the rule of the focused one.
	Border, Focus string
	// Button, ButtonHover and ButtonActive are the backgrounds of a clickable
	// item at rest, under the pointer and pressed (or the selected tab).
	// ButtonHover is a colour between the other two. ButtonText is the text on
	// the first two, ButtonOn the text on the pressed one.
	Button, ButtonHover, ButtonActive, ButtonText, ButtonOn string
	// Colour is false for the plain theme: attributes (bold, reverse) carry
	// the meaning instead.
	Colour bool
}

// Dark suits a dark terminal background.
var Dark = Tokens{
	Name: "dark", Accent: "212", Success: "114", Warning: "215", Error: "203", Info: "75",
	Muted: "245", Faint: "240", Text: "252", Highlight: "238", RowHover: "235", Bar: "236",
	Border: "240", Focus: "212", Colour: true,
	Button: "#2b1948", ButtonHover: "#490ca3", ButtonActive: "#6800ff", ButtonText: "#ffffff", ButtonOn: "#ffffff",
}

// Light suits a light terminal background; the colours keep a contrast of 4.5:1.
var Light = Tokens{
	Name: "light", Accent: "125", Success: "28", Warning: "130", Error: "160", Info: "25",
	Muted: "241", Faint: "243", Text: "235", Highlight: "253", RowHover: "255", Bar: "254",
	Border: "243", Focus: "125", Colour: true,
	Button: "#e6dcf7", ButtonHover: "#b48cf9", ButtonActive: "#6800ff", ButtonText: "#2b1948", ButtonOn: "#ffffff",
}

// Plain has no colour at all (NO_COLOR): bold and reverse video mark the
// selection, and every state also has a symbol and a word.
var Plain = Tokens{Name: "plain"}
