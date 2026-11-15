package derive

import (
	"os"
	"path/filepath"
	"slices"
)

// converge asks collect for the proposals and, when they name a file that is
// neither on disk nor owned yet, asks once more with those files counted as
// owned. A file may depend on another that this run creates (a tsconfig.json
// that references a tsconfig.lib.json): the second answer sees the file, so
// one run reaches the state that a second run would. It returns the last answer.
func converge(root string, owned []string, collect func(owned []string) ([]Proposal, []Divergence, error)) ([]Proposal, []Divergence, error) {
	proposals, diverging, err := collect(owned)
	if err != nil {
		return nil, nil, err
	}
	next := slices.Clone(owned)
	for _, p := range proposals {
		if slices.Contains(next, p.File.Path) {
			continue
		}
		if _, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(p.File.Path))); statErr != nil {
			next = append(next, p.File.Path)
		}
	}
	if len(next) == len(owned) {
		return proposals, diverging, nil
	}
	return collect(next)
}
