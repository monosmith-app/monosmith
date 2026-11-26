package dashboard

import (
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// list draws "Needs you now" with its count, then each item of q on two lines,
// with a blank line between two items on a wide page.
func (c *canvas) list(q []Item) {
	th := c.th
	count := ""
	if c.ch.Item >= 0 {
		count = "  " + th.Muted(words.Say("dashboard.list.count", "at", strconv.Itoa(c.ch.Item+1), "total", strconv.Itoa(len(q))))
	}
	c.add(" "+th.Bold(words.Say("dashboard.list.title"))+count, "")
	if c.wide {
		for _, it := range q {
			c.tagW = max(c.tagW, pane.Width(tagOf(it.Kind)))
		}
	}
	for i, it := range q {
		if i > 0 && c.wide {
			c.add("")
		}
		c.item(i, it)
	}
}

// item draws an item on two lines. The first has the mark in the colour of
// its severity, on a wide page the tag, the title in bold and the actions at
// the right; the second has the detail, muted, under the title. The selected
// item has an accent bar at its left and an accent title.
func (c *canvas) item(i int, it Item) {
	th := c.th
	chosen := i == c.ch.Item
	row := len(c.lines)
	bar := "  "
	if chosen {
		bar = th.Accent("│") + " "
	}
	head := c.icon(it.Severity) + " "
	if c.wide {
		tag := tagOf(it.Kind)
		head += c.paint(it.Severity, tag) + strings.Repeat(" ", c.tagW-pane.Width(tag)+2)
	}
	indent := 1 + 2 + pane.Width(head)
	right := c.actionsWidth(i, it)
	x0 := c.w - 1 - right
	title := pane.CutEllipsis(it.Title, max(x0-indent-2, 1))
	if chosen {
		title = th.Bold(th.Accent(title))
	} else {
		title = th.Bold(title)
	}
	c.targets = append(c.targets, Target{Row: row, From: 0, To: max(x0-1, 1), Kind: TargetItem, Item: i})
	line := pane.Fit(" "+bar+head+title, x0) + c.actions(i, it, row, x0)
	detail := ""
	if it.Detail != "" {
		detail = th.Muted(pane.CutEllipsis(it.Detail, max(c.w-1-indent, 1)))
	}
	c.targets = append(c.targets, Target{Row: row + 1, From: 0, To: c.w, Kind: TargetItem, Item: i})
	c.add(line, " "+bar+strings.Repeat(" ", pane.Width(head))+detail)
	c.items = append(c.items, Rows{From: row, To: row + 2})
}

// actionsWidth is the cells the actions of an item take at the right.
func (c *canvas) actionsWidth(i int, it Item) int {
	if len(it.Buttons) == 0 {
		return 0
	}
	if !c.wide {
		return pane.Width(commands.KeyLabel(c.narrowKey(i, it)))
	}
	w := 0
	for j, b := range it.Buttons {
		if j > 0 {
			w += buttonGap
		}
		w += button.Width(c.th, b.Label, c.keyOf(b, c.picked(i, j)))
	}
	return w
}

// actions draws the actions of item i from column x of line row: its buttons
// on a wide page, the selected one primary, or only the key of its action on
// a narrow page, in the accent.
func (c *canvas) actions(i int, it Item, row, x int) string {
	if len(it.Buttons) == 0 {
		return ""
	}
	if !c.wide {
		key := c.narrowKey(i, it)
		if key == "" {
			return ""
		}
		place := 0
		if i == c.ch.Item {
			place = c.ch.Button
		}
		label := commands.KeyLabel(key)
		c.targets = append(c.targets, Target{Row: row, From: x, To: x + pane.Width(label), Command: it.Buttons[place].CommandID, Kind: TargetKey, Item: i, Button: place})
		return c.th.Accent(label)
	}
	out := ""
	for j, b := range it.Buttons {
		if j > 0 {
			out += strings.Repeat(" ", buttonGap)
			x += buttonGap
		}
		key := c.keyOf(b, c.picked(i, j))
		out += c.draw(b, key, row, x, c.picked(i, j), i, j)
		x += button.Width(c.th, b.Label, key)
	}
	return out
}

// picked is true when the keyboard is on button j of item i.
func (c *canvas) picked(i, j int) bool { return i == c.ch.Item && j == c.ch.Button }

// narrowKey is the key a narrow item shows: Enter on the selected item, else
// the key of its first button ("" when it has none).
func (c *canvas) narrowKey(i int, it Item) string {
	if i == c.ch.Item {
		return c.keyOf(it.Buttons[c.ch.Button], true)
	}
	return c.keyOf(it.Buttons[0], false)
}

// tagOf is the word that says what an item is about.
func tagOf(k Kind) string {
	switch k {
	case KindDoctor, KindSetup, KindRun:
		return words.Say("dashboard.tag.failed")
	case KindWarnings:
		return words.Say("dashboard.tag.warning")
	case KindChanges:
		return words.Say("dashboard.tag.changed")
	case KindRelease:
		return words.Say("dashboard.tag.release")
	case KindUpdates:
		return words.Say("dashboard.tag.updates")
	case KindRemote:
		return words.Say("dashboard.tag.cache")
	case KindCompletion:
		return words.Say("dashboard.tag.setup")
	case KindGood:
		return words.Say("dashboard.tag.good")
	case KindChecking:
		return words.Say("dashboard.tag.checking")
	}
	return words.Say("dashboard.tag.start")
}
