package settingsview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// candidate is the choice the highlighted row would make: the row's theme or
// symbols, the other half as saved. The environment is left out on purpose, so
// the preview shows the choice even where a variable would override it.
func (m Model) candidate(it item) theme.Choice {
	t, i := m.savedTheme, m.savedIcons
	switch it.kind {
	case kindTheme:
		t = normalise(it.key)
	case kindIcons:
		i = normalise(it.key)
	}
	return theme.ChooseWith(func(string) string { return "" }, m.b.Style.Dark, t, i)
}

// previewLines draws a sample of the interface in the highlighted choice, and
// says which environment variable wins over it, if one does.
func (m Model) previewLines(th theme.Theme, it item, w int) []string {
	c := m.candidate(it)
	pt := theme.New(c)
	out := []string{pt.Title("Preview") + th.Muted("  theme "+c.Tokens.Name+", "+iconsName(c))}
	out = append(out, " "+strings.Join([]string{
		pt.Success(pt.Icons.Passed + " passed"), pt.Success(pt.Icons.Cached + " cached"), pt.Failure(pt.Icons.Failed + " failed"),
		pt.Warning(pt.Icons.Warning + " warning"), pt.Info(pt.Icons.Running + " running"), pt.Muted(pt.Icons.Queued + " queued"),
	}, "  "))
	out = append(out, pt.Selected(pane.Fit(" "+pt.Icons.Pointer+" a selected row", min(w, 30)))+"  "+pt.Accent("accent")+" "+pt.Muted("muted")+" "+pt.Faint("faint"))
	out = append(out, " "+progress.Draw(pt, 3, 5, min(max(w/3, 10), 24), false))
	if path := m.b.Style.Path; path != "" {
		out = append(out, th.Muted("Saved for you alone in "+path+"; smith.toml is not touched."))
	}
	getenv := m.b.Style.Getenv
	if getenv == nil {
		return out
	}
	themeVar, iconsVar := theme.Overrides(getenv)
	if themeVar != "" && it.kind == kindTheme {
		out = append(out, th.Warning(themeVar+" is set, so it wins over this choice: unset it to see the effect."))
	}
	if iconsVar != "" && it.kind == kindIcons {
		out = append(out, th.Warning(iconsVar+" is set, so it wins over this choice: unset it to see the effect."))
	}
	return out
}

func iconsName(c theme.Choice) string {
	if c.Icons.Passed == theme.ASCII.Passed {
		return "ascii symbols"
	}
	return "unicode symbols"
}
