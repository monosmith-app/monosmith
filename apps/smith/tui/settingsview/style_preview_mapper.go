package settingsview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
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
	out := []string{pt.Title(words.Say("settingsview.preview.title")) + th.Muted(words.Say("settingsview.preview.theme", "theme", c.Tokens.Name, "symbols", iconsName(c)))}
	out = append(out, " "+strings.Join([]string{
		pt.Success(pt.Icons.Passed + words.Say("settingsview.preview.passed")), pt.Success(pt.Icons.Cached + words.Say("settingsview.preview.cached")), pt.Failure(pt.Icons.Failed + words.Say("settingsview.preview.failed")),
		pt.Warning(pt.Icons.Warning + words.Say("settingsview.preview.warning")), pt.Info(pt.Icons.Running + words.Say("settingsview.preview.running")), pt.Muted(pt.Icons.Queued + words.Say("settingsview.preview.queued")),
	}, "  "))
	out = append(out, pt.Selected(pane.Fit(" "+pt.Icons.Pointer+words.Say("settingsview.preview.selected"), min(w, 30)))+"  "+pt.Accent(words.Say("settingsview.preview.accent"))+" "+pt.Muted(words.Say("settingsview.preview.muted"))+" "+pt.Faint(words.Say("settingsview.preview.faint")))
	out = append(out, " "+progress.Draw(pt, 3, 5, min(max(w/3, 10), 24), false))
	if path := m.b.Style.Path; path != "" {
		out = append(out, th.Muted(words.Say("settingsview.preview.saved", "path", path)))
	}
	getenv := m.b.Style.Getenv
	if getenv == nil {
		return out
	}
	themeVar, iconsVar := theme.Overrides(getenv)
	if themeVar != "" && it.kind == kindTheme {
		out = append(out, th.Warning(words.Say("settingsview.preview.wins", "variable", themeVar)))
	}
	if iconsVar != "" && it.kind == kindIcons {
		out = append(out, th.Warning(words.Say("settingsview.preview.wins", "variable", iconsVar)))
	}
	return out
}

func iconsName(c theme.Choice) string {
	if c.Icons.Passed == theme.ASCII.Passed {
		return words.Say("settingsview.preview.ascii")
	}
	return words.Say("settingsview.preview.unicode")
}
