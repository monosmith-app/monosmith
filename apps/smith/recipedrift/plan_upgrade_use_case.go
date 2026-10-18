package recipedrift

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/recipestack"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PlanUpgrade turns drift reports into the edit of smith.toml that brings
// the projects up to date: their recorded recipe version and answers, the
// settings the recipe gives their plugins and, when the recipe now resolves
// to other plugins, the stack declaration they share. It changes nothing.
//
// A project is left as it is, with the reason, when its stack is provided by
// a plugin, or is shared with a project the recipe does not change the same
// way (including one made without a recipe): rewriting the stack would change
// that project too.
func PlanUpgrade(s *session.Session, reports []Report) (Plan, error) {
	raw, err := os.ReadFile(filepath.Join(s.Root, workspace.FileName))
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{Before: string(raw), After: string(raw)}
	groups := map[string][]Report{}
	var order []string
	for _, r := range reports {
		switch {
		case r.Problem != "":
			plan.Skipped = append(plan.Skipped, r.Summary())
		case r.Drifted():
			name := r.Stack.Name
			if _, ok := groups[name]; !ok {
				order = append(order, name)
			}
			groups[name] = append(groups[name], r)
		}
	}
	var decls []recipestack.Declaration
	var replace []string
	for _, name := range order {
		members := groups[name]
		reason := stackProblem(s, name, members)
		if reason != "" {
			for _, m := range members {
				plan.Skipped = append(plan.Skipped, fmt.Sprintf("%s is not upgraded: %s", m.Project, reason))
			}
			continue
		}
		old, _ := s.Stacks.Get(name)
		rewrite := !sameStack(old, members[0].Stack)
		if rewrite {
			replace = append(replace, name)
		}
		for i, m := range members {
			decls = append(decls, recipestack.Declaration{
				Project: m.Project, Stack: members[0].Stack, Reuse: !(rewrite && i == 0), Settings: m.Settings,
				Recipe: workspace.ProjectRecipe{Name: m.Recipe, Version: m.To, Source: m.Ref, Answers: m.Answers},
			})
			plan.Changes = append(plan.Changes, Change{Project: m.Project, Recipe: m.Recipe, From: m.From, To: m.To, Changes: m.Changes})
			if len(m.Added) > 0 {
				plan.Starters = append(plan.Starters, fmt.Sprintf("%s: the recipe now adds %s; the files a new project gets from a plugin are starter files, which an upgrade does not write (`smith doctor` lists what is missing)", m.Project, strings.Join(m.Added, ", ")))
			}
		}
	}
	if len(decls) == 0 {
		return plan, nil
	}
	after, err := recipestack.Rewrite(plan.Before, decls, replace)
	if err != nil {
		return plan, err
	}
	plan.After = after
	return plan, nil
}

// stackProblem says why the stack the members share cannot be rewritten for
// the recipe, or "" when it can, or when it needs no rewriting.
func stackProblem(s *session.Session, name string, members []Report) string {
	old, err := s.Stacks.Get(name)
	if err != nil {
		return err.Error()
	}
	now := members[0].Stack
	rewrite := false
	for _, m := range members {
		if !sameStack(m.Stack, now) {
			return fmt.Sprintf("the recipe resolves the projects that share the stack %s to different plugins", name)
		}
		rewrite = rewrite || !sameStack(old, m.Stack)
	}
	if !rewrite {
		return ""
	}
	if _, declared := s.Config.Stacks[name]; !declared {
		return fmt.Sprintf("the stack %s is provided by a plugin, not declared in smith.toml, so smith cannot change the plugins it holds", name)
	}
	in := map[string]bool{}
	for _, m := range members {
		in[m.Project] = true
	}
	var others []string
	for _, p := range s.Graph.Projects {
		if p.Stack == name && !in[p.Name] {
			others = append(others, p.Name)
		}
	}
	if len(others) > 0 {
		sort.Strings(others)
		return fmt.Sprintf("the stack %s is shared with %s, which the recipe does not change the same way, and changing its plugins would change them too", name, strings.Join(others, ", "))
	}
	if err := pluginhost.ValidateStack(now, s.Plugins); err != nil {
		return fmt.Sprintf("the recipe's plugins cannot form the stack %s: %v", name, err)
	}
	if s.Plugins.Get(now.Language) == nil {
		return fmt.Sprintf("the recipe uses the language plugin %s, which is not enabled in this workspace", now.Language)
	}
	return ""
}

// sameStack compares the plugins of two stacks, ignoring name, description
// and the order of a list.
func sameStack(a, b protocol.Stack) bool {
	return a.Language == b.Language && a.Scaffold == b.Scaffold &&
		sameSet(a.Toolchain, b.Toolchain) && sameSet(a.Structure, b.Structure) && sameSet(a.Deploy, b.Deploy)
}

func sameSet(a, b []string) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)
	return slices.Equal(x, y)
}
