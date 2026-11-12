package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme renders text in the meaning of a token. It is a value: pass it down to
// every component that draws.
type Theme struct {
	Tokens Tokens
	Icons  Icons

	accent, success, warning, failure, info, muted, faint, text lipgloss.Style
	selected, hover, bar, title                                 lipgloss.Style
	button, buttonHover, buttonActive                           lipgloss.Style
}

// New builds a theme from a choice.
func New(c Choice) Theme {
	t := c.Tokens
	fg := func(col string) lipgloss.Style {
		if !t.Colour || col == "" {
			return lipgloss.NewStyle()
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(col))
	}
	bg := func(col string) lipgloss.Style {
		if !t.Colour || col == "" {
			return lipgloss.NewStyle()
		}
		return lipgloss.NewStyle().Background(lipgloss.Color(col))
	}
	th := Theme{Tokens: t, Icons: c.Icons,
		accent: fg(t.Accent).Bold(true), success: fg(t.Success), warning: fg(t.Warning),
		failure: fg(t.Error).Bold(true), info: fg(t.Info), muted: fg(t.Muted), faint: fg(t.Faint),
		text: fg(t.Text), title: lipgloss.NewStyle().Bold(true),
		bar: bg(t.Bar).Inherit(fg(t.Text)),
	}
	if t.Colour {
		th.selected = bg(t.Highlight).Bold(true).Foreground(lipgloss.Color(t.Accent))
		th.hover = bg(t.RowHover)
		th.button = bg(t.Button).Foreground(lipgloss.Color(t.ButtonText))
		th.buttonHover = bg(t.ButtonHover).Foreground(lipgloss.Color(t.ButtonText))
		th.buttonActive = bg(t.ButtonActive).Foreground(lipgloss.Color(t.ButtonOn)).Bold(true)
	} else {
		th.selected = lipgloss.NewStyle().Reverse(true).Bold(true)
		th.hover = lipgloss.NewStyle().Underline(true)
		th.buttonHover = lipgloss.NewStyle().Reverse(true)
		th.buttonActive = lipgloss.NewStyle().Reverse(true).Bold(true)
	}
	return th
}

// Default is the dark theme with Unicode symbols.
func Default() Theme { return New(Choice{Tokens: Dark, Icons: Unicode}) }

// Paint renders s in an exact foreground and, when given, background colour
// (#rrggbb). Without colour it is s unchanged.
func (t Theme) Paint(s, fg, bg string) string {
	if !t.Tokens.Colour {
		return s
	}
	style := lipgloss.NewStyle()
	if fg != "" {
		style = style.Foreground(lipgloss.Color(fg))
	}
	if bg != "" {
		style = style.Background(lipgloss.Color(bg))
	}
	return style.Render(s)
}

// Accent, Success, Warning, Failure, Info, Muted, Faint, Title, Selected,
// Hover and Bar render s in the token's meaning.
func (t Theme) Accent(s string) string   { return t.accent.Render(s) }
func (t Theme) Success(s string) string  { return t.success.Render(s) }
func (t Theme) Warning(s string) string  { return t.warning.Render(s) }
func (t Theme) Failure(s string) string  { return t.failure.Render(s) }
func (t Theme) Info(s string) string     { return t.info.Render(s) }
func (t Theme) Muted(s string) string    { return t.muted.Render(s) }
func (t Theme) Faint(s string) string    { return t.faint.Render(s) }
func (t Theme) Title(s string) string    { return t.title.Render(s) }
func (t Theme) Selected(s string) string { return t.selected.Render(s) }
func (t Theme) Hover(s string) string    { return t.hover.Render(s) }

// Button, ButtonHover and ButtonActive render s as a clickable item at rest,
// under the pointer, and pressed or selected. Without colour a resting item is
// told apart by its brackets (the caller adds them, see the button package) and
// the other two by reverse video.
func (t Theme) Button(s string) string       { return t.button.Render(s) }
func (t Theme) ButtonHover(s string) string  { return t.buttonHover.Render(s) }
func (t Theme) ButtonActive(s string) string { return t.buttonActive.Render(s) }

// Bold renders s in bold.
func (t Theme) Bold(s string) string { return lipgloss.NewStyle().Bold(true).Render(s) }

// Border colours the rule of a pane: the focused pane's is the accent.
func (t Theme) Border(s string, focused bool) string {
	if focused {
		return t.accent.Render(s)
	}
	return t.faint.Render(s)
}

// Spin is the spinner frame for a tick count.
func (t Theme) Spin(tick int) string {
	f := t.Icons.Spinner
	return f[((tick%len(f))+len(f))%len(f)]
}

// Bar renders s on the bar background. Styled segments inside s end with a
// reset that would drop the background, so the background is put back after
// each one.
func (t Theme) Bar(s string) string { return under(t.bar, s) }

// RowHover renders a whole row (a list row, an option, a tick box row) in the
// hover look: on the RowHover background, kept under the styled segments the
// row already has. Without colour it is underlined. Every row under the
// pointer takes this one look; a button takes ButtonHover.
func (t Theme) RowHover(s string) string { return under(t.hover, s) }

// under renders s in style, and puts the style back after every reset inside
// s, so a styled segment does not end it.
func under(style lipgloss.Style, s string) string {
	open, reset, ok := strings.Cut(style.Render("@"), "@")
	if !ok || open == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\x1b[0m", reset+open)
	s = strings.ReplaceAll(s, "\x1b[m", reset+open)
	return open + s + reset
}
