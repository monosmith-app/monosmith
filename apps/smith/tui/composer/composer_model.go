package composer

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/layout"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// stage is the step the screen is on.
type stage int

const (
	// stBrowse: search the catalogue and tick items.
	stBrowse stage = iota
	// stQuestion: a tick left an axis open; the person answers it.
	stQuestion
	// stName: the project is named and its kind chosen.
	stName
	// stPlan: the plan is shown, with what to do with it.
	stPlan
)

// kinds are the kinds of project the composer can add, in the order Tab cycles.
var kinds = []string{"app", "lib", "package"}

// projectName is the shape of a project name smith add accepts.
var projectName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Model is the "Compose a stack" screen: a search box over the catalogue, the
// ticked items grouped by layer, the live check of the ticks, and the plan.
// It holds the catalogue and checks ticks with the solver of smith add, so it
// asks its host for nothing until the plan.
type Model struct {
	cat    compose.Catalogue
	offers []choiceselect.Offer

	search textinput.Model
	name   textinput.Model
	hits   []compose.Hit
	cursor int
	// hover is the row under the pointer (of the list, or of the question's
	// answers), -1 for none.
	hover int
	// cursorLine is the drawn line of the cursor row, so the list scrolls to it.
	cursorLine int
	// scroll is where the list's window starts; a pointer, so the wheel can move
	// it between frames without moving the cursor (see pane.Scroll).
	scroll *pane.Scroll

	// lang is the language item ticked, if any; ticked are the choice items
	// ticked, in the order they were ticked.
	lang   *compose.Item
	ticked []compose.Item
	chk    compose.Check

	// stack shows the ticked items by layer instead of the catalogue.
	stack bool
	// onList is true while the keyboard is on the list, not in the search box:
	// Space then ticks the row under the cursor instead of typing a space.
	onList  bool
	stage   stage
	qcursor int
	kind    int
	plan    *compose.Plan
	planErr string
	note    string

	// top is the first line shown of a question, the name or the plan when
	// the text is taller than the screen.
	top int

	w, h int
	rect layout.Rect
	// margins are the rows above and below the screen that stay free.
	margins layout.Margins
}

// New builds the screen on a catalogue (the output of smith compose --json).
//
// Parameters: cat is the catalogue.
// Returns: the screen with nothing ticked, searching for everything.
func New(cat compose.Catalogue) Model {
	search := textinput.New()
	search.Placeholder = words.Say("composer.search.placeholder")
	search.Focus()
	name := textinput.New()
	name.Placeholder = words.Say("composer.name.placeholder")
	m := Model{cat: cat, offers: compose.OffersOf(cat.Items), search: search, name: name, scroll: &pane.Scroll{}, hover: -1}
	return m.refresh()
}

// WithSearch starts the screen with words in the search box.
//
// Parameters: words are the words to search for.
func (m Model) WithSearch(words string) Model {
	m.search.SetValue(words)
	return m.refresh()
}

// Within keeps the dialog off the top rows (the header and the tab strip) and
// the bottom rows (the footer and the status bar) of the screen, as every
// dialog does: Place centres it in the rows between.
//
// Parameters: top and bottom are the rows to leave free.
// Returns: the dialog with the margins set.
func (m Model) Within(top, bottom int) Model {
	m.margins = layout.Margins{Top: max(top, 0), Bottom: max(bottom, 0)}
	return m
}

// Place centres the screen on a terminal of sw by sh cells, between the margins
// Within set.
func (m Model) Place(sw, sh int) Model {
	m.w, m.h = sw, sh
	w := min(max(sw-6, 40), 104)
	m.rect = m.margins.Centre(sw, sh, w, 34)
	m.search.Width = max(w-14, 10)
	m.name.Width = max(w-20, 10)
	return m.follow(m.stage)
}

// Wheel scrolls the list, or the text of a question, the name or the plan, by
// delta lines; the cursor stays where it is.
func (m Model) Wheel(delta int) Model {
	if m.stage != stBrowse {
		return m.scrollText(delta)
	}
	m.scroll.By(delta)
	return m
}

// Rect is where the screen sits after Place.
func (m Model) Rect() layout.Rect { return m.rect }

// Check is what the ticks add up to now.
func (m Model) Check() compose.Check { return m.chk }

// Ticked are the names ticked, in order.
func (m Model) Ticked() []string {
	out := make([]string, len(m.ticked))
	for i, it := range m.ticked {
		out[i] = it.Name
	}
	return out
}

// WithPlan gives the screen the plan its host read, or the reason it could not.
//
// Parameters: plan is the plan; err is why there is none, or nil.
func (m Model) WithPlan(plan compose.Plan, err error) Model {
	m.stage, m.top = stPlan, 0
	if err != nil {
		m.plan, m.planErr = nil, err.Error()
		return m
	}
	m.plan, m.planErr = &plan, ""
	return m
}

// language is the language the ticks share: the ticked language item, else the
// language of the first ticked choice, else empty.
func (m Model) language() string {
	switch {
	case m.lang != nil:
		return m.lang.Name
	case len(m.ticked) > 0:
		return m.ticked[0].Language
	}
	return ""
}

// refresh finds the items for the search box, hides those of another language,
// and checks the ticks again.
func (m Model) refresh() Model {
	lang := m.language()
	var visible []compose.Hit
	for _, h := range compose.Search(m.cat.Items, m.search.Value()) {
		if lang != "" && h.Kind != compose.KindLanguage && h.Language != "" && h.Language != lang {
			continue
		}
		visible = append(visible, h)
	}
	m.hits = visible
	m.cursor = min(m.cursor, max(m.length()-1, 0))
	pickLang := ""
	if m.lang != nil {
		pickLang = m.lang.Name
	}
	m.chk = compose.CheckItems(m.offers, pickLang, m.ticked)
	if len(m.ticked) == 0 {
		m.chk = compose.Check{Language: pickLang, Missing: words.Say("composer.missing.empty")}
	}
	return m
}

// length is the number of rows the cursor moves over.
func (m Model) length() int {
	if m.stack {
		return len(m.ticked) + b2i(m.lang != nil)
	}
	return len(m.hits)
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// tick ticks or unticks the item under the cursor. A workspace plugin is not
// ticked: the note says which command turns it on.
func (m Model) tick(it compose.Item) Model {
	m.note = ""
	switch {
	case it.Kind == compose.KindPlugin:
		m.note = words.Say("composer.note.plugin", "title", it.Title, "route", it.Route)
		return m
	case it.Problem != "":
		m.note = words.Say("composer.note.problem", "title", it.Title, "problem", it.Problem)
		return m
	case it.Kind == compose.KindLanguage:
		if m.lang != nil && m.lang.ID == it.ID {
			m.lang = nil
		} else {
			picked := it
			m.lang = &picked
		}
	default:
		if i := slices.IndexFunc(m.ticked, func(t compose.Item) bool { return t.ID == it.ID }); i >= 0 {
			m.ticked = slices.Delete(slices.Clone(m.ticked), i, i+1)
		} else {
			// A single-choice axis acts like radio buttons: the new pick takes
			// the place of the one before, and the note says so.
			rivals := compose.Rivals(m.ticked, it)
			kept := slices.DeleteFunc(slices.Clone(m.ticked), func(t compose.Item) bool {
				return slices.ContainsFunc(rivals, func(r compose.Item) bool { return r.ID == t.ID })
			})
			m.ticked = append(kept, it)
			if len(rivals) > 0 {
				names := make([]string, len(rivals))
				for i, r := range rivals {
					names[i] = r.Title
				}
				m.note = words.Say("composer.note.replaces", "title", it.Title, "names", strings.Join(names, words.Say("composer.and")), "axis", it.Axis, "first", names[0])
			}
		}
	}
	m = m.refresh()
	if m.chk.Question != nil && m.stage == stBrowse {
		m.stage, m.qcursor = stQuestion, suggestedIndex(m.chk.Question)
	}
	return m
}

// suggestedIndex is the option a question starts on: the suggested one.
func suggestedIndex(q *compose.Question) int {
	for i, o := range q.Options {
		if o.Suggested {
			return i
		}
	}
	return 0
}

// Key handles a key. A screen that is newly open starts from its top; a
// question follows its cursor.
func (m Model) Key(msg tea.KeyMsg) (Model, Outcome) {
	was := m.stage
	m, out := m.key(msg)
	return m.follow(was), out
}

// key is Key before the text follows the cursor.
func (m Model) key(msg tea.KeyMsg) (Model, Outcome) {
	if m.stage != stBrowse {
		page := pane.PageStep(m.rect.H - 4)
		switch {
		case msg.String() == "ctrl+u" || commands.NavNamed(msg.String()) == commands.NavPageUp:
			return m.scrollText(-page), Outcome{}
		case msg.String() == "ctrl+d" || commands.NavNamed(msg.String()) == commands.NavPageDown:
			return m.scrollText(page), Outcome{}
		}
	}
	switch m.stage {
	case stQuestion:
		return m.questionKey(msg.String())
	case stName:
		return m.nameKey(msg)
	case stPlan:
		return m.planKey(msg.String())
	}
	return m.browseKey(msg)
}

// browseKey handles a key while the person searches and ticks. The arrows put
// the keyboard on the list, where Space ticks, as in every list; a letter puts
// it back in the search box. Enter goes on, as Ctrl+N does.
func (m Model) browseKey(msg tea.KeyMsg) (Model, Outcome) {
	last := max(m.length()-1, 0)
	switch commands.NavNamed(msg.String()) {
	case commands.NavUp:
		if m.cursor == 0 {
			// Up from the first row goes back to the search box.
			m.onList = false
			return m.focusSearch(), Outcome{}
		}
		m.cursor--
		return m.focusList(), Outcome{}
	case commands.NavDown:
		if m.onList {
			m.cursor = min(m.cursor+1, last)
		}
		// From the search box the first Down lands on the row under the cursor.
		return m.focusList(), Outcome{}
	case commands.NavPageUp:
		m.cursor = max(m.cursor-8, 0)
		return m.focusList(), Outcome{}
	case commands.NavPageDown:
		m.cursor = min(m.cursor+8, last)
		return m.focusList(), Outcome{}
	case commands.NavHome:
		m.cursor = 0
		return m.focusList(), Outcome{}
	case commands.NavEnd:
		m.cursor = last
		return m.focusList(), Outcome{}
	}
	switch msg.String() {
	case "ctrl+c":
		return m, Outcome{Done: true}
	case "esc":
		switch {
		case m.search.Value() != "":
			m.search.SetValue("")
			return m.focusSearch().refresh(), Outcome{}
		case m.stack:
			m.stack, m.cursor = false, 0
			return m.focusSearch(), Outcome{}
		}
		return m, Outcome{Done: true}
	case "tab":
		m.stack, m.cursor = !m.stack, 0
	case " ":
		if m.onList {
			return m.enter(), Outcome{}
		}
		return m.typeIn(msg), Outcome{}
	case "enter", "ctrl+n":
		return m.next()
	default:
		return m.typeIn(msg), Outcome{}
	}
	return m, Outcome{}
}

// typeIn gives a key to the search box, which takes the keyboard back.
func (m Model) typeIn(msg tea.KeyMsg) Model {
	m = m.focusSearch()
	before := m.search.Value()
	m.search, _ = m.search.Update(msg)
	if m.search.Value() != before {
		m.cursor = 0
		m = m.refresh()
	}
	return m
}

// focusList puts the keyboard on the list: the search box shows no cursor.
func (m Model) focusList() Model {
	m.onList = true
	m.search.Blur()
	return m
}

// focusSearch puts the keyboard in the search box.
func (m Model) focusSearch() Model {
	m.onList = false
	m.search.Focus()
	return m
}

// enter ticks the item under the cursor, or in the stack view unticks it.
func (m Model) enter() Model {
	if m.stack {
		entries := m.entries()
		if m.cursor < len(entries) {
			m = m.tick(entries[m.cursor])
			m.cursor = min(m.cursor, max(m.length()-1, 0))
		}
		return m
	}
	if m.cursor < len(m.hits) {
		return m.tick(m.hits[m.cursor].Item)
	}
	return m
}

// entries are the ticked items as the stack view lists them: the language
// first, then the choices by layer.
func (m Model) entries() []compose.Item {
	var out []compose.Item
	if m.lang != nil {
		out = append(out, *m.lang)
	}
	rest := slices.Clone(m.ticked)
	slices.SortStableFunc(rest, func(a, b compose.Item) int { return layerIndex(m.cat, a.Layer) - layerIndex(m.cat, b.Layer) })
	return append(out, rest...)
}

// layerIndex is the place of a layer in the catalogue's layers.
func layerIndex(cat compose.Catalogue, id string) int {
	for i, l := range cat.Layers {
		if l.ID == id {
			return i
		}
	}
	return len(cat.Layers)
}

// next moves on from browsing: to the open question, or to the name when the
// ticks make a project. A conflict or a missing piece stays on the screen.
func (m Model) next() (Model, Outcome) {
	switch {
	case m.chk.Conflict != "":
		m.note = words.Say("composer.note.conflict", "conflict", m.chk.Conflict)
	case m.chk.Question != nil:
		m.stage, m.qcursor = stQuestion, suggestedIndex(m.chk.Question)
	case !m.chk.Complete:
		m.note = firstNonEmpty(m.chk.Missing, words.Say("composer.note.incomplete"))
	default:
		m.stage = stName
		m.name.Focus()
		m.search.Blur()
	}
	return m, Outcome{}
}

// questionKey handles a key while an open axis is asked.
func (m Model) questionKey(key string) (Model, Outcome) {
	q := m.chk.Question
	if q == nil {
		m.stage = stBrowse
		return m, Outcome{}
	}
	switch commands.NavNamed(key) {
	case commands.NavCancel:
		m.stage = stBrowse
	case commands.NavUp:
		m.qcursor = max(m.qcursor-1, 0)
	case commands.NavDown:
		m.qcursor = min(m.qcursor+1, len(q.Options)-1)
	case commands.NavConfirm:
		answer := q.Options[min(m.qcursor, len(q.Options)-1)]
		for _, it := range m.cat.Items {
			if it.Kind == compose.KindChoice && it.Name == answer.Name && it.Language == m.language() {
				m.ticked = append(slices.Clone(m.ticked), it)
				break
			}
		}
		m = m.refresh()
		if m.chk.Question != nil {
			m.qcursor = suggestedIndex(m.chk.Question)
		} else {
			m.stage = stBrowse
		}
	}
	return m, Outcome{}
}

// nameKey handles a key while the project is named.
func (m Model) nameKey(msg tea.KeyMsg) (Model, Outcome) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.stage = stBrowse
		m.name.Blur()
		m = m.focusSearch()
		m.note = ""
	case "tab":
		m.kind = (m.kind + 1) % len(kinds)
	case "enter":
		value := strings.TrimSpace(m.name.Value())
		if !projectName.MatchString(value) {
			m.note = words.Say("composer.note.name")
			return m, Outcome{}
		}
		m.note = ""
		m.stage = stPlan
		m.plan, m.planErr = nil, ""
		return m, Outcome{Plan: &PlanRequest{Kind: kinds[m.kind], Name: value, Language: m.chk.Language, Picks: m.chk.Selection}}
	default:
		m.name, _ = m.name.Update(msg)
	}
	return m, Outcome{}
}

// planKey handles a key while the plan is shown.
func (m Model) planKey(key string) (Model, Outcome) {
	switch key {
	case "ctrl+c":
		return m, Outcome{Done: true}
	case "esc", "b":
		m.stage = stName
	case "r":
		if m.plan != nil && len(m.plan.Argv) > 1 {
			return m, Outcome{Done: true, Run: m.plan.Argv[1:]}
		}
	case "c":
		if m.plan != nil {
			return m, Outcome{Copy: m.plan.Command, Note: words.Say("composer.note.copied")}
		}
	case "s":
		if m.plan != nil && len(m.plan.RecipeArgv) > 1 {
			return m, Outcome{Done: true, Run: m.plan.RecipeArgv[1:]}
		}
	default:
		switch commands.NavNamed(key) {
		case commands.NavUp:
			return m.scrollText(-1), Outcome{}
		case commands.NavDown:
			return m.scrollText(1), Outcome{}
		}
	}
	return m, Outcome{}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
