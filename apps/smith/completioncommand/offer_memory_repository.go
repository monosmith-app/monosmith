package completioncommand

import (
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/userprefs"
)

// The answers that are remembered.
const (
	// ChoiceNo means the person said no: smith never asks again.
	ChoiceNo = "no"
	// ChoiceLater means the person wants to be asked again after AskAfter.
	ChoiceLater = "later"
	// ChoiceInstalled means smith wrote the completion script.
	ChoiceInstalled = "installed"
)

// Answer is what the person answered to the offer of tab completion.
type Answer struct {
	// Choice is ChoiceNo, ChoiceLater, ChoiceInstalled or "" (not asked yet).
	Choice string
	// AskAfter is when to ask again; it only counts for ChoiceLater.
	AskAfter time.Time
}

// Memory keeps the answer between runs of smith.
type Memory interface {
	// Load reads the answer; a missing or broken store is "not asked yet".
	Load() Answer
	// Save keeps the answer.
	Save(Answer) error
}

// PreferencesMemory keeps the answer in the preferences of the person, the
// file tui.json in the smith home (SMITH_HOME, or ~/.smith). It is a personal
// choice, so it never goes into the smith.toml of a workspace.
type PreferencesMemory struct {
	// Home is the smith home; "" keeps nothing.
	Home string
}

// Load reads the answer from the preferences.
func (m PreferencesMemory) Load() Answer {
	p, _ := userprefs.Load(m.Home)
	return Answer{Choice: p.CompletionOffer, AskAfter: p.CompletionAskAfter}
}

// Save writes the answer into the preferences and leaves the other
// preferences as they are.
func (m PreferencesMemory) Save(a Answer) error {
	p, _ := userprefs.Load(m.Home)
	p.CompletionOffer, p.CompletionAskAfter = a.Choice, a.AskAfter
	return userprefs.Save(m.Home, p)
}
