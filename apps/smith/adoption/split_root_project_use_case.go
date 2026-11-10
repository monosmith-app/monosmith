package adoption

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/adoptsplit"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// SplitRootProject cuts the files of the project that adoptmove.Apply just
// moved, as its language declares: the tools go to the workspace manifest,
// the shared configuration to the root, and the project keeps the rest. A
// language with no split does nothing. It stages the result (the tree was
// clean before the move, so nothing else is staged) and tells the person
// each change.
//
// Parameters: r is the report; c the project; p the plan that was applied.
// Returns: an error when a file cannot be read or written; git reset --hard
// puts the tree back.
func SplitRootProject(r Report, c adoptmove.Candidate, p adoptmove.Plan, out io.Writer) error {
	split := adoptsplit.For(c.Ecosystem).Split
	if split == nil {
		return nil
	}
	res, err := split(adoptsplit.Request{
		Root: r.Root, Dir: p.Dir, Workspace: r.Name,
		Patterns: workspacePatterns([]string{p.Dir}),
		Hoist:    p.Hoist, Drop: p.Replaced, KeepLint: p.KeepLint,
	})
	if err != nil {
		return err
	}
	if res.ESLint {
		if err := AddESLintConfig(r.Root, out); err != nil {
			return err
		}
	}
	if _, err := gitclient.New(r.Root).Run("add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	for _, note := range res.Notes {
		fmt.Fprintf(out, "  split: %s\n", note)
	}
	return nil
}
