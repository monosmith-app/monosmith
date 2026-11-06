package tui

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/tui/composer"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// What the composer's captured commands were asked for.
const (
	capCompose     = "compose"
	capComposePlan = "compose-plan"
)

// composeReport is the part of `smith compose --json` the screen reads.
type composeReport struct {
	Layers []compose.Layer `json:"layers"`
	Items  []compose.Item  `json:"items"`
	Plan   *compose.Plan   `json:"plan"`
}

// openCompose opens the "Compose a stack" screen. The catalogue is read the
// first time, which takes a moment, and kept.
func (s *Shell) openCompose() tea.Cmd {
	if s.composeCat != nil {
		return s.showCompose()
	}
	argv := []string{"compose", "--json"}
	return s.captureWorking(capCompose, words.Say("tui.compose.compose_a_stack"), words.Say("tui.compose.the_screen_lists_what_your_plugins"), argv, argv, words.Say("tui.compose.reading_what_your_plugins_offer"), words.Say("tui.compose.cancelled_the_screen_was_not_opened"))
}

// composeRead keeps the catalogue and opens the screen; when it cannot be read
// the person is told why.
func (s *Shell) composeRead(msg capturedMsg) tea.Cmd {
	var rep composeReport
	if !msg.res.OK || json.Unmarshal([]byte(msg.res.Text), &rep) != nil {
		s.notify(words.Say("tui.compose.could_not_read_the_catalogue", "firstLine", firstLine(failureText(msg.res))), toast.Warning)
		return nil
	}
	s.composeCat = &compose.Catalogue{Layers: rep.Layers, Items: rep.Items}
	return s.showCompose()
}

// showCompose draws the screen on the catalogue.
func (s *Shell) showCompose() tea.Cmd {
	m := composer.New(*s.composeCat).Place(s.w, s.h)
	s.compose = &m
	return nil
}

// composePlanRead gives the screen the plan the command printed.
func (s *Shell) composePlanRead(msg capturedMsg) tea.Cmd {
	if s.compose == nil {
		return nil
	}
	var rep composeReport
	var err error
	switch {
	case !msg.res.OK:
		err = errText(firstLine(failureText(msg.res)))
	case json.Unmarshal([]byte(msg.res.Text), &rep) != nil || rep.Plan == nil:
		err = errText(words.Say("tui.compose.smith_did_not_print_a_plan"))
	}
	m := s.compose.WithPlan(planOrEmpty(rep.Plan), err)
	s.compose = &m
	return nil
}

type errText string

func (e errText) Error() string { return string(e) }

func planOrEmpty(p *compose.Plan) compose.Plan {
	if p == nil {
		return compose.Plan{}
	}
	return *p
}

// composeKey sends a key to the screen and does what it asks.
func (s *Shell) composeKey(msg tea.KeyMsg) tea.Cmd {
	m, out := s.compose.Key(msg)
	s.compose = &m
	switch {
	case out.Plan != nil:
		return s.capture(capComposePlan, out.Plan.Argv())
	case out.Copy != "":
		s.copyText(out.Copy)
	case out.Run != nil:
		s.compose = nil
		return s.execute(out.Run, false)
	}
	if out.Done {
		s.compose = nil
	}
	return nil
}
