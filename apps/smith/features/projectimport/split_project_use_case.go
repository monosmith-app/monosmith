package projectimport

import (
	"fmt"
	"io"
	"path"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoption"
	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/features/adoptsplit"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
)

// The values of --eslint-setup.
const (
	ESLintShared = "shared"
	ESLintKeep   = "keep"
)

// SplitAnswers are how the questions of the split are answered.
type SplitAnswers struct {
	// ESLint is --eslint-setup: ESLintShared, ESLintKeep or "" to ask.
	ESLint string
	// Hoist is --hoist: development dependencies that go to the workspace
	// without a question.
	Hoist []string
	// Ask is true when a terminal can answer.
	Ask bool
	In  io.Reader
	Out io.Writer
}

// Split cuts the files of the project that now sits in its folder, the way
// adopt cuts the project it moves (adoptsplit): the tools among its
// development dependencies go to the workspace manifest, the shared
// configuration to the root, and the project keeps the rest. A development
// dependency smith does not know as a tool, and the project's own ESLint
// setup, are asked about in a terminal; without one the project keeps them
// and smith says so. A language with no split does nothing.
//
// Parameters: root is the workspace root; p is the plan; a says how to answer.
// Returns: the notes for the person, or an error from a file or a question.
func Split(root string, p Plan, a SplitAnswers) ([]string, error) {
	rules := adoptsplit.For(p.Ecosystem)
	if rules.Split == nil {
		return nil, nil
	}
	dir := filepath.Join(root, filepath.FromSlash(p.Folder))
	var sorted adoptsplit.Sorted
	if rules.Sort != nil {
		var err error
		if sorted, err = rules.Sort(dir); err != nil {
			return nil, err
		}
	}
	var notes []string
	hoist := append([]string{}, sorted.Tools...)
	for _, name := range sorted.Unsure {
		if slices.Contains(a.Hoist, name) {
			hoist = append(hoist, name)
			continue
		}
		if !a.Ask {
			notes = append(notes, name+" stays in the project's own development dependencies (--hoist "+name+" sends it to the workspace)")
			continue
		}
		toWorkspace, err := adoptmove.AskWorkspace(name, a.In, a.Out)
		if err != nil {
			return notes, err
		}
		if toWorkspace {
			hoist = append(hoist, name)
		}
	}
	keepLint, err := lintChoice(sorted.Lint, a)
	if err != nil {
		return notes, err
	}
	var drop []string
	if !keepLint {
		drop = sorted.Replaced
	}
	res, err := rules.Split(adoptsplit.Request{
		Root: root, Dir: p.Folder, Workspace: filepath.Base(root),
		Patterns: []string{path.Dir(p.Folder) + "/*"},
		Hoist:    hoist, Drop: drop, KeepLint: keepLint, InPlace: true,
	})
	if err != nil {
		return notes, err
	}
	if res.ESLint {
		if err := adoption.AddESLintConfig(root, a.Out); err != nil {
			return notes, err
		}
	}
	return append(notes, res.Notes...), nil
}

// lintChoice settles whether the project keeps its own ESLint setup: the
// flag, else the answer in a terminal, else keep (it changes the least).
func lintChoice(has bool, a SplitAnswers) (bool, error) {
	if !has {
		return true, nil
	}
	switch a.ESLint {
	case ESLintShared:
		return false, nil
	case ESLintKeep:
		return true, nil
	case "":
		if a.Ask {
			return adoptmove.AskLint(a.In, a.Out)
		}
		return true, nil
	}
	return true, problem.New(fmt.Sprintf("--eslint-setup %s is not a choice", a.ESLint), "use --eslint-setup shared or --eslint-setup keep")
}
