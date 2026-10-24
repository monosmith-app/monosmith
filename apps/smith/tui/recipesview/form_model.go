package recipesview

import (
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Mode is what the form creates.
type Mode int

// The things the form can create.
const (
	NewProject Mode = iota
	NewWorkspace
	NewRecipe
)

var modeNames = []string{"new project", "new workspace", "new recipe"}

// DefaultRecipePath is where the recipe wizard writes unless told otherwise.
const DefaultRecipePath = "team.recipe.toml"

// Model is the Recipes view: a form built from the recipe's own choices, with
// what the recipe puts in each role shown before anything is written. What
// the form builds is always a smith command line, shown under it:
//
//	new project    smith add <kind> <name> --choice name=answer...
//	new workspace  smith new <name> --recipe <recipe>
//	new recipe     smith recipe new <path>   (the wizard, in a terminal of its own)
type Model struct {
	Loaded  bool
	Catalog Catalog
	Mode    Mode
	Kind    string
	// Name is the name of the project or workspace, Path where the recipe wizard writes.
	Name string
	Path string
	// Recipe is the recipe a new workspace is started with ("" for none).
	Recipe string
	// Answers are the choices the user changed.
	Answers map[string]string
	Detail  Detail
	// Cursor is the row selected.
	Cursor int

	editing string // the row being typed into, "" when none
	input   textinput.Model
	// detailKey says what Detail describes, pending what is being loaded.
	detailKey string
	pending   string
	// Problem is feedback for the last key (a field that cannot change).
	Problem string
}

// New is an empty form that waits for its catalog.
func New() Model {
	return Model{Kind: Kinds[0], Path: DefaultRecipePath, Answers: map[string]string{}}
}

// FromLoaded is the form once the catalog arrived.
func (v Model) FromLoaded(l Loaded) Model {
	v.Loaded, v.Catalog = true, l.Catalog
	v.detailKey, v.pending, v.Detail = "", "", Detail{}
	return v
}

// Reset forgets the catalog and what was described, keeping what was typed.
func (v Model) Reset() Model {
	v.Loaded, v.Catalog, v.Detail, v.detailKey, v.pending = false, Catalog{}, Detail{}, "", ""
	return v
}

// Editing is true while a text field has the keyboard.
func (v Model) Editing() bool { return v.editing != "" }

// row is one line of the form.
type row struct {
	id string
	// question is the choice a choice row sets.
	question Question
}

// NoRecipes is true when the workspace has no recipe at all (no *.recipe.toml
// and none named in smith.toml): there is nothing to create a project or a
// workspace from, so the form offers only to write the first recipe.
func (v Model) NoRecipes() bool {
	if !v.Loaded || v.Catalog.Err != nil || len(v.Catalog.Recipes) > 0 {
		return false
	}
	for _, ref := range v.Catalog.ByKind {
		if ref != "" {
			return false
		}
	}
	return true
}

// mode is what the form creates: with no recipe in the workspace, only a recipe.
func (v Model) mode() Mode {
	if v.NoRecipes() {
		return NewRecipe
	}
	return v.Mode
}

// rows are the form's fields for the mode and the recipe's choices.
func (v Model) rows() []row {
	var out []row
	if !v.NoRecipes() {
		out = append(out, row{id: "mode"})
	}
	switch v.mode() {
	case NewProject:
		out = append(out, row{id: "kind"}, row{id: "name"})
		// The choices of the last description stay while an answer reloads it.
		if v.ref() != "" && v.Detail.Ref == v.ref() {
			for _, q := range v.Detail.Questions {
				out = append(out, row{id: "choice", question: q})
			}
		}
	case NewWorkspace:
		out = append(out, row{id: "name"}, row{id: "recipe"})
	case NewRecipe:
		out = append(out, row{id: "path"})
	}
	return out
}

// ref is the recipe the form describes now.
func (v Model) ref() string {
	switch v.mode() {
	case NewProject:
		return v.Catalog.ByKind[v.Kind]
	case NewWorkspace:
		return v.Recipe
	}
	return ""
}

// key says what the form wants described: the recipe and the answers.
func (v Model) key() string {
	names := make([]string, 0, len(v.Answers))
	for n := range v.Answers {
		names = append(names, n+"="+v.Answers[n])
	}
	sort.Strings(names)
	return v.ref() + "\x00" + strings.Join(names, "\x00")
}

func (v Model) freshDetail() bool { return v.ref() != "" && v.detailKey == v.key() }

// Request says whether the recipe must be described again, and for what; the
// shell loads it in the background, since a recipe can be a URL.
func (v Model) Request() (key, ref string, answers map[string]string, ok bool) {
	ref = v.ref()
	key = v.key()
	if !v.Loaded || ref == "" || key == v.detailKey || key == v.pending {
		return "", "", nil, false
	}
	answers = map[string]string{}
	for n, a := range v.Answers {
		answers[n] = a
	}
	return key, ref, answers, true
}

// Requested records that the description for key is being loaded.
func (v Model) Requested(key string) Model {
	v.pending = key
	return v
}

// Receive takes a description that arrived; one for answers the form has
// moved on from is dropped.
func (v Model) Receive(msg DetailLoaded) Model {
	if v.pending == msg.Key {
		v.pending = ""
	}
	if msg.Key != v.key() {
		return v
	}
	v.Detail, v.detailKey = msg.Detail, msg.Key
	v.Cursor = min(v.Cursor, len(v.rows())-1)
	return v
}

// Update handles a key.
func (v Model) Update(key tea.KeyMsg) Model {
	v.Problem = ""
	if v.editing != "" {
		return v.typing(key)
	}
	rows := v.rows()
	v.Cursor = min(max(v.Cursor, 0), len(rows)-1)
	switch key.String() {
	case "up", "k":
		v.Cursor = max(v.Cursor-1, 0)
	case "down", "j":
		v.Cursor = min(v.Cursor+1, len(rows)-1)
	case "home":
		v.Cursor = 0
	case "end":
		v.Cursor = len(rows) - 1
	case "left", "h":
		return v.change(-1)
	case "right", "l":
		return v.change(1)
	case "enter", " ":
		return v.activate()
	}
	return v
}

// typing sends a key to the text field being edited.
func (v Model) typing(key tea.KeyMsg) Model {
	switch key.Type {
	case tea.KeyEsc:
		v.editing = ""
	case tea.KeyEnter:
		text := strings.TrimSpace(v.input.Value())
		switch {
		case v.editing == "name":
			v.Name = text
		case v.editing == "path":
			v.Path = text
		case strings.HasPrefix(v.editing, "choice:"):
			name := strings.TrimPrefix(v.editing, "choice:")
			v.Answers = withAnswer(v.Answers, name, text)
		}
		v.editing = ""
	default:
		v.input, _ = v.input.Update(key)
	}
	return v
}

// activate is Enter on the selected row: a text field is edited, a choice moves to its next answer.
func (v Model) activate() Model {
	rows := v.rows()
	if v.Cursor >= len(rows) {
		return v
	}
	r := rows[v.Cursor]
	switch {
	case r.id == "name":
		return v.edit("name", v.Name)
	case r.id == "path":
		return v.edit("path", v.Path)
	case r.id == "choice" && len(r.question.Options) == 0:
		return v.edit("choice:"+r.question.Name, r.question.Answer)
	}
	return v.change(1)
}

// edit starts typing into a field.
func (v Model) edit(field, value string) Model {
	v.editing = field
	v.input = textinput.New()
	v.input.SetValue(value)
	v.input.Focus()
	return v
}

// change moves the selected row's value to the next or previous one.
func (v Model) change(dir int) Model {
	rows := v.rows()
	if v.Cursor >= len(rows) {
		return v
	}
	r := rows[v.Cursor]
	switch r.id {
	case "mode":
		v.Mode = Mode((int(v.Mode) + dir + len(modeNames)) % len(modeNames))
	case "kind":
		v.Kind = cycle(Kinds, v.Kind, dir)
	case "recipe":
		v.Recipe = cycle(v.recipeOptions(), v.Recipe, dir)
	case "choice":
		if len(r.question.Options) > 0 {
			v.Answers = withAnswer(v.Answers, r.question.Name, cycle(r.question.Options, r.question.Answer, dir))
		}
	}
	v.Cursor = min(v.Cursor, len(v.rows())-1)
	return v
}

// recipeOptions are what a new workspace can start from: no recipe, or a
// recipe file of this workspace that reads as a recipe.
func (v Model) recipeOptions() []string {
	out := []string{""}
	for _, l := range v.Catalog.Recipes {
		if l.Error == "" {
			out = append(out, l.Path)
		}
	}
	return out
}

// cycle is the option after (or before) current, wrapping; an unknown current starts at the first.
func cycle(options []string, current string, dir int) string {
	if len(options) == 0 {
		return current
	}
	at := slices.Index(options, current)
	if at < 0 {
		return options[0]
	}
	return options[(at+dir+len(options))%len(options)]
}

// withAnswer is the answers with one set, or removed when the text is empty.
func withAnswer(answers map[string]string, name, value string) map[string]string {
	out := map[string]string{}
	for n, a := range answers {
		out[n] = a
	}
	if value == "" {
		delete(out, name)
	} else {
		out[name] = value
	}
	return out
}

// SelectRow moves the cursor to row i; a row that was already selected acts, as Enter does.
func (v Model) SelectRow(i int) Model {
	if v.editing != "" || i < 0 || i >= len(v.rows()) {
		return v
	}
	if v.Cursor == i {
		return v.activate()
	}
	v.Cursor = i
	return v
}

// Argv is the smith command line the form builds, or why it cannot yet: the
// reason is shown beside the button that would run it.
func (v Model) Argv() (argv []string, reason string) {
	switch v.mode() {
	case NewRecipe:
		if v.Path == "" {
			return nil, "type where to write the recipe (Enter edits the path)"
		}
		return []string{"recipe", "new", v.Path}, ""
	case NewWorkspace:
		if v.Name == "" {
			return nil, "type the workspace's name (Enter edits it)"
		}
		argv = []string{"new", v.Name}
		if v.Recipe != "" {
			argv = append(argv, "--recipe", v.Recipe)
		}
		return argv, ""
	}
	if v.ref() == "" {
		return nil, "no recipe makes " + v.Kind + "s: set one in smith.toml ([workspace] recipe or [recipes] " + v.Kind + "), or use Add project in the Guide"
	}
	switch {
	case v.Name == "":
		return nil, "type the project's name (Enter edits it)"
	case v.Detail.Problem != "" && v.freshDetail():
		return nil, "the recipe has a problem (see above)"
	case !v.freshDetail():
		return nil, "reading the recipe…"
	case len(v.Detail.Unanswered) > 0:
		return nil, "answer " + strings.Join(v.Detail.Unanswered, ", ") + " first (left and right change a choice)"
	}
	argv = []string{"add", v.Kind, v.Name}
	for _, q := range v.Detail.Questions {
		if q.Given {
			argv = append(argv, "--choice", q.Name+"="+q.Answer)
		}
	}
	return argv, ""
}
