package tui

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/tui/projectlist"
	"github.com/monosmith-app/monosmith/libs/plugin-docker/images"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The catalogue actions whose form is built from the command's description:
// the buttons of the detail open that form.
const (
	formKind     = "kind"
	formEject    = "eject"
	formTauriAdd = "tauri add" // words:literal
)

// projectFormActions are the buttons under More in the detail that open a form
// about the selected project, each shown only when it applies to it: the swap
// of a stack choice, the environments of a web app, the Tauri forms of a Node
// app and the Docker Hub descriptions of an app that ships as an image.
func (s Shell) projectFormActions(p projectlist.Project) []detail.Action {
	var out []detail.Action
	if p.Stack != "" {
		out = append(out, s.commandButton("proj.swap"))
	}
	if isWebApp(p) {
		out = append(out, s.commandButton("ws.env-add"), s.commandButton("ws.env-remove"))
	}
	if isTauriApp(p) {
		out = append(out, s.commandButton("ws.tauri-icon"))
	} else if canAddTauri(p) && s.hasForm(formTauriAdd) {
		out = append(out, detail.Action{ID: "x:tauri-add", Label: words.Say("tui.detail.make_it_a_tauri_app"), Hint: words.Say("tui.detail.make_it_a_tauri_app_hint"), Disabled: s.running()})
	}
	if shipsImage(p) {
		out = append(out, s.commandButton("ws.docker-description"))
	}
	return out
}

// projectMoreActions are the last buttons under More in the detail of a
// project: what changes the project as a whole, used rarely. Each one asks
// before it changes anything.
func (s Shell) projectMoreActions() []detail.Action {
	var out []detail.Action
	if s.hasForm(formKind) {
		out = append(out, detail.Action{ID: "x:kind", Label: words.Say("tui.detail.change_kind"), Hint: words.Say("tui.detail.change_kind_hint"), Disabled: s.running()})
	}
	if s.hasForm(formEject) {
		out = append(out, detail.Action{ID: "x:eject", Label: words.Say("tui.detail.eject"), Hint: words.Say("tui.detail.eject_hint"), Disabled: s.running()})
	}
	return append(out, s.commandButton("proj.remove"))
}

// commandButton is the detail button of a palette command, with the
// command's own title and hint, so the button and the palette entry read the same.
func (s Shell) commandButton(id string) detail.Action {
	c, _ := commands.ByID(id)
	a := s.action(id, c.Title)
	a.Hint = c.Hint
	return a
}

// hasForm reports whether the menu has the form of a catalogue action.
func (s Shell) hasForm(id string) bool {
	_, ok := s.guide.ActionPresenting(id)
	return ok
}

// isWebApp reports whether environments apply to a project: an app written in Node.
func isWebApp(p projectlist.Project) bool { return p.Kind == "app" && p.Language == "node" }

// isTauriApp reports whether a project is a Tauri desktop app: it has the task
// that builds the installers.
func isTauriApp(p projectlist.Project) bool {
	_, ok := p.Task(tauri.TaskName)
	return ok
}

// shipsImage reports whether a project ships as a Docker image: it has the
// task that builds the image.
func shipsImage(p projectlist.Project) bool {
	_, ok := p.Task(images.TaskName)
	return ok
}
