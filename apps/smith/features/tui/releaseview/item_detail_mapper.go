package releaseview

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// DetailTitle is the title of the detail: the selected package and its
// versions ("web 1.0.0 → 1.1.0"), or "" when no package is selected.
func (v Model) DetailTitle() string {
	item, ok := v.Selected()
	if !ok {
		return ""
	}
	from := item.Current
	if from == "" {
		from = words.Say("releaseview.new")
	}
	return words.Say("releasepage.detail.title", "project", item.Project, "from", from, "to", item.Next)
}

// Detail tells about the selected package: its bump, why it releases, the
// commits behind the bump, its tag, what the public API check found and the
// command line that shows the same.
//
// Parameters: th draws.
// Returns: the sections of the detail in order; none when no package is selected.
func (v Model) Detail(th theme.Theme) []detail.Section {
	item, ok := v.Selected()
	if !ok {
		return nil
	}
	out := []detail.Section{
		{Items: []string{BumpMark(th, item.Level) + th.Muted(" · "+words.Say("releasepage.detail.plugin", "plugin", item.Plugin))}},
		{Heading: words.Say("releasepage.detail.why"), Items: []string{item.Reason}},
		{Heading: words.Say("releasepage.detail.changes"), Items: v.commits(th, item)},
		{Heading: words.Say("releasepage.detail.tag"), Items: []string{item.Tag}},
	}
	if len(item.APIWarnings) > 0 {
		var warn []string
		for _, w := range item.APIWarnings {
			warn = append(warn, th.Warning(th.Icons.Warning+" "+w))
		}
		out = append(out, detail.Section{Heading: words.Say("releasepage.detail.api"), Items: warn})
	}
	return append(out, detail.Section{Heading: words.Say("releasepage.detail.same_as"), Items: []string{th.Faint(words.Say("releasepage.detail.command"))}, OneLine: true})
}
