package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/recipestack"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// selectionPlan is a selection of choices (smith add --with) solved for one new
// project, before anything is written.
type selectionPlan struct {
	given     []string
	selection choiceselect.Selection
	stack     protocol.Stack
	// reuse is true when a stack with the same choices exists: nothing is
	// declared then.
	reuse bool
}

// planSelection solves the choices of smith add --with against what the
// loaded plugins offer and settles the stack the project gets: an existing
// stack with the same choices, else a new one named after them. It writes
// nothing. It returns nil when --with is not given.
func planSelection(s *session.Session, opts AddOptions) (*selectionPlan, error) {
	if len(opts.With) == 0 {
		return nil, nil
	}
	if opts.Stack != "" {
		return nil, fmt.Errorf("--with names the choices and --stack names a stack: use one or the other")
	}
	if opts.Recipe != nil && len(opts.Recipe.Answers) > 0 {
		return nil, fmt.Errorf("--with names the choices and --choice answers a recipe: use one or the other")
	}
	sel, err := choiceselect.Solve(s.Plugins.Offers(), opts.With)
	if err != nil {
		return nil, err
	}
	st := sel.Stack("", "A selection of choices: "+strings.Join(sel.Names(), ", "))
	if opts.Lang != "" {
		asked, err := s.Plugins.ForLanguage(opts.Lang)
		if err != nil {
			return nil, err
		}
		if asked != s.Plugins.Get(st.Language) {
			return nil, fmt.Errorf("--lang %s, but the choices %s are for %s projects", opts.Lang, strings.Join(opts.With, ", "), st.Language)
		}
	}
	plan := &selectionPlan{given: opts.With, selection: sel}
	for _, known := range s.Stacks.All() {
		if known.Problem == "" && sameChoices(known.Choices, st.Choices) {
			plan.stack, plan.reuse = known.Stack, true
			return plan, nil
		}
	}
	name, reuse, err := recipestack.Choose(strings.Join(opts.With, "-"), st, func(n string) (protocol.Stack, bool) {
		for _, known := range s.Stacks.All() {
			if known.Name == n {
				return known.Stack, true
			}
		}
		return protocol.Stack{}, false
	})
	if err != nil {
		return nil, err
	}
	st.Name = name
	plan.stack, plan.reuse = st, reuse
	return plan, nil
}

// describe is the line smith add prints for the selection: the choices, the
// defaults it added and the stack they make.
func (p *selectionPlan) describe() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Choices: %s", strings.Join(p.selection.Names(), ", "))
	if len(p.selection.Added) > 0 {
		fmt.Fprintf(&b, " (%s by default)", strings.Join(p.selection.Added, ", "))
	}
	fmt.Fprintf(&b, " as the stack %s", p.stack.Name)
	return b.String()
}

// declare saves the stack in smith.toml as [stacks.<name>] with its choices
// and reloads the session from it. It returns the function that puts
// smith.toml back when creating the project fails. A reused stack needs no
// declaration, and its undo does nothing.
func (p *selectionPlan) declare(s *session.Session) (func(), error) {
	if p.reuse {
		return func() {}, nil
	}
	path := filepath.Join(s.Root, workspace.FileName)
	original, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(original)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	quoted := make([]string, len(p.stack.Choices))
	for i, c := range p.stack.Choices {
		quoted[i] = configedit.Quote(c)
	}
	text += fmt.Sprintf("\n# Written by smith from smith add --with %s.\n[stacks.%s]\ndescription = %s\nchoices = [%s]\n",
		strings.Join(p.given, ","), p.stack.Name, configedit.Quote(p.stack.Description), strings.Join(quoted, ", "))
	if _, err := workspace.Parse([]byte(text)); err != nil {
		return nil, fmt.Errorf("the stack for --with %s would break smith.toml: %w", strings.Join(p.given, ","), err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return nil, err
	}
	rollback := func() {
		if os.WriteFile(path, original, 0o644) == nil {
			_ = s.Reload()
		}
	}
	if err := s.Reload(); err != nil {
		rollback()
		return nil, fmt.Errorf("smith.toml with the new stack does not load: %w", err)
	}
	return rollback, nil
}

// sameChoices reports whether two lists hold the same names, in any order.
func sameChoices(a, b []string) bool {
	if len(a) == 0 || len(a) != len(b) {
		return false
	}
	a, b = slices.Sorted(slices.Values(a)), slices.Sorted(slices.Values(b))
	return slices.Equal(a, b)
}
