package swapchoice

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Prepared is a swap that smith can carry out: both toolchain plugins, what
// they add, and what moving from one to the other changes.
type Prepared struct {
	Swap      Swap
	Project   *projectgraph.Project
	Migration Migration
	// Stack is the stack the project has afterwards, "" when no stack matches.
	Stack      string
	params     protocol.GenerateParams
	next       tooling.Setup
	root       string
	fromPlugin string
}

// PrepareMigration works out the files a swap changes. It reads and writes
// nothing. The axes smith can migrate by itself are the ones whose choices a
// Node toolchain plugin owns and whose setup is mechanical: the test runner,
// Jest and Vitest.
//
// Parameters: s is the open session; p is the project; swap is the checked swap.
// Returns: the prepared swap; a problem that says what is left to do by hand
// when the axis or the plugins are not migrated by smith.
func PrepareMigration(s *session.Session, p *projectgraph.Project, swap Swap) (Prepared, error) {
	offers := s.Plugins.Offers()
	from, to := offerNamed(offers, swap.From), offerNamed(offers, swap.To)
	oldPlugin, okOld := toolchain(from.Plugin)
	newPlugin, okNew := toolchain(to.Plugin)
	_, fromKnown := testModules[from.Plugin]
	_, toKnown := testModules[to.Plugin]
	if swap.Axis != "test" || !okOld || !okNew || !fromKnown || !toKnown {
		return Prepared{}, problem.New(
			fmt.Sprintf("smith cannot move %s from %s to %s by itself yet: only the test runner (jest and vitest) has a mechanical migration", p.Name, swap.From, swap.To),
			"change its dependencies, config files and imports by hand, then run smith doctor to check the result",
			"or make a new project with smith add --with "+strings.Join(swap.Choices, ",")+" and move the code over")
	}
	facts, err := projectgraph.ConfigFacts(s.Config, p.Name)
	if err != nil {
		return Prepared{}, err
	}
	params := protocol.GenerateParams{
		Kind: p.Kind, Name: p.Name, Directory: p.Root, Stack: p.Stack, Choices: swap.Choices, Facts: facts, Shape: p.Shape,
		Workspace: protocol.WorkspaceInfo{Name: s.Config.Workspace.Name, Scope: s.Config.Workspace.Scope, ModulePrefix: s.Config.Workspace.ModulePrefix},
	}
	oldSetup, newSetup := oldPlugin.Setup(s.Root, params), newPlugin.Setup(s.Root, params)
	shared := false
	for _, name := range s.Graph.Names() {
		if other := s.Graph.Get(name); other.Name != p.Name && other.Stack != "" {
			if st, err := s.Stacks.Get(other.Stack); err == nil && slices.Contains(st.Plugins(), from.Plugin) {
				shared = true
			}
		}
	}
	m, err := PlanMigration(s.Root, p.Root, oldSetup, newSetup, testModules[from.Plugin], testModules[to.Plugin], shared)
	if err != nil {
		return Prepared{}, err
	}
	if shared && len(m.Kept) > 0 {
		m.ByHand = append(m.ByHand, "other projects still use "+swap.From+", so its packages stay in the root package.json: "+strings.Join(m.Kept, ", "))
	}
	return Prepared{Swap: swap, Project: p, Migration: m, Stack: stackAfter(s, p, from.Plugin, to.Plugin), params: params, next: newSetup, root: s.Root, fromPlugin: from.Plugin}, nil
}

// Lines describe the migration for the plan, one line each.
func (pr Prepared) Lines() []string {
	var out []string
	add := func(verb string, names []string) {
		if len(names) > 0 {
			out = append(out, fmt.Sprintf("  %s: %s", verb, strings.Join(names, ", ")))
		}
	}
	m := pr.Migration
	add("removes from the root package.json devDependencies", m.Dependencies)
	add("removes package.json scripts", m.Scripts)
	add("removes package.json fields", m.Fields)
	add("deletes files", m.Files)
	add("moves the test imports in", m.RewriteNames())
	var adds []string
	adds = append(adds, sortedKeys(pr.next.DevDependencies)...)
	add("adds devDependencies if absent", adds)
	add("adds scripts if absent", sortedKeys(pr.next.Scripts))
	add("writes files if absent", sortedKeys(pr.next.Files))
	if pr.Stack != "" && pr.Stack != pr.Project.Stack {
		out = append(out, "  records the stack "+pr.Stack+" in package.json")
	}
	return out
}

// Apply carries the prepared swap out: it removes what the old plugin wrote,
// has the new plugin write its own, moves the test imports, records the new
// stack, installs the packages once, and runs the project's build and tests.
//
// Parameters: ctx stops the run; s is the open session; out receives progress.
// Returns: what is left to do by hand, or the error that stopped the swap. A
// verification that fails returns an error too; the files stay changed.
func (pr Prepared) Apply(ctx context.Context, s *session.Session, out io.Writer) ([]string, error) {
	project := filepath.Join(pr.root, filepath.FromSlash(pr.Project.Root))
	m := pr.Migration
	fmt.Fprintf(out, "Moving %s from %s to %s\n", pr.Project.Name, pr.Swap.From, pr.Swap.To)
	if err := editJSON(filepath.Join(pr.root, "package.json"), func(doc *jsondoc.Document) {
		if deps := doc.Object("devDependencies", false); deps != nil {
			for _, name := range m.Dependencies {
				deps.Delete(name)
			}
		}
	}); err != nil {
		return nil, err
	}
	if err := editJSON(filepath.Join(project, "package.json"), func(doc *jsondoc.Document) {
		if scripts := doc.Object("scripts", false); scripts != nil {
			for _, name := range m.Scripts {
				scripts.Delete(name)
			}
		}
		for _, name := range m.Fields {
			doc.Delete(name)
		}
	}); err != nil {
		return nil, err
	}
	for _, name := range m.Files {
		if err := os.Remove(filepath.Join(project, filepath.FromSlash(name))); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	for name, body := range m.Rewrites {
		if err := os.WriteFile(filepath.Join(project, filepath.FromSlash(name)), []byte(body), 0o644); err != nil {
			return nil, err
		}
	}
	fmt.Fprintf(out, "… writing what %s needs\n", pr.Swap.To)
	res, err := tooling.Apply(pr.root, pr.params, pr.next)
	if err != nil {
		return nil, fmt.Errorf("writing the files of %s: %w", pr.Swap.To, err)
	}
	if pr.Stack != "" && pr.Stack != pr.Project.Stack {
		if err := editJSON(filepath.Join(project, "package.json"), func(doc *jsondoc.Document) {
			doc.Object("smith", true).SetString("stack", pr.Stack)
		}); err != nil {
			return nil, err
		}
	}
	if res.InstallPending {
		if err := install(ctx, s, out); err != nil {
			return m.ByHand, err
		}
	}
	if err := s.Refresh(); err != nil {
		return m.ByHand, err
	}
	if p := s.Graph.Get(pr.Project.Name); p != nil {
		if err := scaffold.VerifyProject(ctx, s, p, out); err != nil {
			return m.ByHand, fmt.Errorf("the files are changed, but %w", err)
		}
	}
	return m.ByHand, nil
}

// install has the Node plugin install the workspace once.
func install(ctx context.Context, s *session.Session, out io.Writer) error {
	node, err := s.Plugins.ForLanguage("node")
	if err != nil {
		return err
	}
	if !node.Supports(protocol.MethodInstall) {
		return fmt.Errorf("the node plugin cannot install packages: run your package manager yourself")
	}
	fmt.Fprintln(out, "Installing packages")
	var installed protocol.DependencyResult
	if err := node.Call(protocol.MethodInstall, protocol.InstallParams{Quiet: true}, &installed, 20*time.Minute); err != nil {
		return fmt.Errorf("installing packages: %w", err)
	}
	return ctx.Err()
}

// stackAfter finds the stack the project has once the old toolchain plugin is
// replaced by the new one: the same language and scaffold, the toolchain list
// with the plugin swapped. It returns "" when the catalog has no such stack.
func stackAfter(s *session.Session, p *projectgraph.Project, from, to string) string {
	current, err := s.Stacks.Get(p.Stack)
	if err != nil {
		return ""
	}
	want := slices.Clone(current.Toolchain)
	for i, name := range want {
		if name == from {
			want[i] = to
		}
	}
	for _, c := range s.Stacks.All() {
		if c.Problem == "" && c.Language == current.Language && c.Scaffold == current.Scaffold && slices.Equal(c.Toolchain, want) {
			return c.Name
		}
	}
	return ""
}

// offerNamed finds the offer with a name; the zero offer when there is none.
func offerNamed(offers []choiceselect.Offer, name string) choiceselect.Offer {
	for _, o := range offers {
		if o.Name == name {
			return o
		}
	}
	return choiceselect.Offer{}
}

// toolchain is the built-in toolchain plugin of a name.
func toolchain(name string) (*tooling.Toolchain, bool) {
	ctor, ok := pluginhost.Builtin(name)
	if !ok {
		return nil, false
	}
	t, ok := ctor("").(*tooling.Toolchain)
	return t, ok
}

// editJSON rewrites a JSON file through edit, keeping its formatting.
func editJSON(path string, edit func(*jsondoc.Document)) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	edit(doc)
	return os.WriteFile(path, doc.Bytes(), 0o644)
}
