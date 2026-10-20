package recipesview

import (
	"github.com/monosmith-app/monosmith/apps/smith/recipecatalog"
)

// Kinds are the project kinds smith add creates, in the order the form offers them.
var Kinds = []string{"app", "lib", "package"}

// Catalog is what the workspace has to offer: the recipe files it holds and
// which recipe each kind of new project is made from.
type Catalog struct {
	Recipes []recipecatalog.Listed
	// ByKind is the recipe ref new projects of each kind are made from ("" for
	// none: the stack or the language plugin alone makes them).
	ByKind map[string]string
	Err    error
}

// Role is a level or slot a recipe fills, and where the recipe said so.
type Role struct {
	Key     string
	Plugins []string
	// Origin is the recipe file and line, relative to the workspace.
	Origin string
}

// Question is one choice the recipe asks, as the form shows it.
type Question struct {
	Name    string
	Ask     string
	Options []string
	Default string
	// Answer is what the choice is set to now: the given answer, else the
	// default, else empty.
	Answer string
	// Given says the answer is the user's, not the default.
	Given bool
}

// Detail is what a recipe offers for the answers given: its choices, and
// before anything is written, what it puts in each role.
type Detail struct {
	// Ref is the recipe described; empty when the project is not made from one.
	Ref     string
	Name    string
	Version string
	// Questions are the choices asked for these answers, in the order the recipe asks them.
	Questions []Question
	// Roles are what the recipe puts in each role; empty while a choice has no answer.
	Roles []Role
	// Configures names the plugins whose settings the recipe sets.
	Configures []string
	// Offers are the levels and slots the recipe may fill, with every plugin
	// it could put there; they say what a choice changes before it is answered.
	Offers []Offer
	// Unanswered names the choices that have neither an answer nor a default.
	Unanswered []string
	// Problem says why the recipe could not be resolved (a broken recipe, not
	// merely an unanswered choice).
	Problem string
}

// Offer is a level or slot with the plugins a recipe puts there always, or
// only for an answer ("language=node").
type Offer struct {
	Key          string
	Plugins      []string
	Alternatives map[string][]string
}

// Loaded is the catalog read in the background: the shell receives it as a message.
type Loaded struct {
	Catalog Catalog
}

// DetailLoaded is a recipe described for some answers; Key says for which, so
// an answer that arrives late is ignored.
type DetailLoaded struct {
	Key    string
	Detail Detail
}
