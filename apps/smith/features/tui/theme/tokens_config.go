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
	// the first two, ButtonOn the text on the pressed one. Button with
	// ButtonText is the secondary button (the Highlight background with the
	// Text colour). ButtonActive with ButtonOn, in bold, is the primary button
	// (the Accent background).
	Button, ButtonHover, ButtonActive, ButtonText, ButtonOn string
	// Colour is false for the plain theme: attributes (bold, reverse) carry
	// the meaning instead.
	Colour bool
}

// Dark suits a dark terminal background. It is the MonoSmith palette (see
// docs/architecture/tui-redesign.md): every text colour keeps a contrast of
// 4.5:1 on the terminal's background, the selected row and the bars.
var Dark = Tokens{
	Name: "dark", Accent: "#5fb6d1", Success: "#7ccf96", Warning: "#e3c16f", Error: "#ff9a90", Info: "#5fb6d1",
	Muted: "#b0b0b0", Faint: "#a5a5a5", Text: "#d0d0d0", Highlight: "#2f3d44", RowHover: "#263238", Bar: "#303030",
	Border: "#5a5a5a", Focus: "#5fb6d1", Colour: true,
	Button: "#2f3d44", ButtonHover: "#3d5560", ButtonActive: "#5fb6d1", ButtonText: "#d0d0d0", ButtonOn: "#141414",
}

// Light suits a light terminal background, in the same palette and with the
// same contrast rule as Dark. Success is #1b7537, a step darker than the
// palette's #1d7a3a, which has only 4.4:1 on the selected row.
var Light = Tokens{
	Name: "light", Accent: "#176a86", Success: "#1b7537", Warning: "#855600", Error: "#b3261e", Info: "#176a86",
	Muted: "#5c5c5c", Faint: "#686868", Text: "#2b2b2b", Highlight: "#dcebf1", RowHover: "#eef5f8", Bar: "#f2f2f2",
	Border: "#b8b8b8", Focus: "#176a86", Colour: true,
	Button: "#dcebf1", ButtonHover: "#a9cfdd", ButtonActive: "#176a86", ButtonText: "#2b2b2b", ButtonOn: "#ffffff",
}

// Plain has no colour at all (NO_COLOR): bold and reverse video mark the
// selection, and every state also has a symbol and a word.
var Plain = Tokens{Name: "plain"}
