package theme

import "github.com/charmbracelet/lipgloss"

// Theme renders text in the meaning of a token. It is a value: pass it down to
// every component that draws.
type Theme struct {
	Tokens Tokens
	Icons  Icons

	accent, success, warning, failure, info, muted, faint, text lipgloss.Style
	selected, hover, bar, title                                 lipgloss.Style
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
		text: fg(t.Text), title: fg(t.Accent).Bold(true),
		bar: bg(t.Bar).Inherit(fg(t.Text)),
	}
	if t.Colour {
		th.selected = bg(t.Highlight).Bold(true).Foreground(lipgloss.Color(t.Accent))
		th.hover = bg(t.Highlight)
	} else {
		th.selected = lipgloss.NewStyle().Reverse(true).Bold(true)
		th.hover = lipgloss.NewStyle().Underline(true)
	}
	return th
}

// Default is the dark theme with Unicode symbols.
func Default() Theme { return New(Choice{Tokens: Dark, Icons: Unicode}) }

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
func (t Theme) Bar(s string) string      { return t.bar.Render(s) }

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
