package projectimport

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/features/manifestoffer"
	"github.com/monosmith-app/monosmith/apps/smith/features/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/features/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Options are the flags and the terminal of smith import.
type Options struct {
	// Source is the git URL or the folder.
	Source string
	// As, Into, Name, Ref, History, ESLint, Manifest and Path are the flags
	// of the same names (ESLint is --eslint-setup); "" when not given.
	As, Into, Name, Ref, History, ESLint, Manifest, Path string
	// Hoist is --hoist.
	Hoist []string
	// Tooling is --tooling: smith or keep (or <project>=smith|keep), the
	// answer of the tooling question of a Node project.
	Tooling []string
	// DryRun shows the plan and changes nothing.
	DryRun bool
	// Yes applies the plan without the question.
	Yes bool
	// Version is smith's.
	Version string
	// Env is the terminal state.
	Env operationplan.Environment
	In  io.Reader
	Out io.Writer
	Log io.Writer
}

// Import brings one project into the workspace that holds the current
// folder: fetch, identify, settle the folder and the history, show the plan
// and ask, then place, split, join, write the pipelines, install and check.
// It refuses, and changes nothing, outside a workspace, for a folder that
// exists, for a source with no project at its top, and for a git history
// on a work tree with uncommitted changes.
//
// Parameters: o are the options.
// Returns: the plan (its Name is "" when nothing was settled), and an error
// that says what stopped the import and the step that clears it.
func Import(o Options) (Plan, error) {
	// One reader for every question: two would each swallow the other's input.
	if o.In != nil {
		o.In = bufio.NewReader(o.In)
	}
	root, err := workspace.FindRoot(".")
	if err != nil {
		return Plan{}, problem.New("this folder is not in a smith workspace, so there is nothing to import into",
			"smith adopt makes this repository a workspace; run smith import in it next", "or cd into a smith workspace")
	}
	if o.ESLint != "" && o.ESLint != ESLintShared && o.ESLint != ESLintKeep {
		return Plan{}, problem.New("--eslint-setup "+o.ESLint+" is not a choice", "use --eslint-setup shared or --eslint-setup keep")
	}
	if _, ok := o.In.(*bufio.Reader); !ok && o.In != nil {
		// One reader for every question and the confirmation: two would each
		// swallow the other's input.
		o.In = bufio.NewReader(o.In)
	}
	claims, taken, err := readWorkspace(root, o.Version, o.Log)
	if err != nil {
		return Plan{}, err
	}
	src, err := Fetch(o.Source, o.Ref, o.Out)
	if err != nil {
		return Plan{}, err
	}
	defer func() { Release(src) }()
	if o.Path != "" {
		if src, err = Narrow(src, o.Path, o.Out); err != nil {
			return Plan{}, err
		}
	}
	p, err := settle(root, &src, claims, taken, o)
	if err != nil {
		return p, err
	}
	if err := SettleTooling(src, &p, &o); err != nil {
		return p, err
	}
	fmt.Fprint(o.Out, Summary(p))
	ok, err := operationplan.Present(o.Out, o.In, Changes(p), operationplan.Options{Yes: o.Yes, DryRun: o.DryRun, Ask: true, Env: o.Env})
	if err != nil || !ok {
		return p, err
	}
	if err := Place(root, p, o.Out); err != nil {
		return p, err
	}
	if p.WriteManifest != "" {
		if err := writeManifest(root, o.Version, &p, o.Out); err != nil {
			return p, fmt.Errorf("the project is in %s, but its manifest was not written (%w); write it yourself, then run smith doctor", p.Folder, err)
		}
	}
	keptLeft := KeepVersions(root, p, o.Out)
	if dropped, err := DropLockfiles(root, p.Folder, Lockfiles(p.Ecosystem)); err != nil {
		return p, err
	} else if len(dropped) > 0 {
		fmt.Fprintf(o.Out, "  split: dropped %s: the workspace has one lockfile\n", strings.Join(dropped, ", "))
	}
	notes, err := Split(root, p, SplitAnswers{ESLint: o.ESLint, Hoist: o.Hoist, Ask: operationplan.ShouldAsk(o.Env, o.Yes), In: o.In, Out: o.Out})
	for _, n := range notes {
		fmt.Fprintf(o.Out, "  split: %s\n", n)
	}
	if err != nil {
		return p, fmt.Errorf("the project is in %s, but the split of workspace and project failed (%w); git status shows what changed", p.Folder, err)
	}
	left, err := Join(root, o.Version, p, o.Out)
	if err != nil {
		return p, err
	}
	left = append(left, keptLeft...)
	left = append(left, ApplyTooling(root, p, o)...)
	installed := Install(root, o.Version, p.Name, o.Out)
	left = append(left, installed...)
	if len(left) == 0 {
		left = append(left, Verify(root, p.Name, o.Out)...)
	}
	if len(installed) == 0 {
		left = append(left, Lint(root, p.Name, o.Out)...)
	}
	left = append(left, LayoutLeft(root, p)...)
	checked, err := Check(root, o.Version)
	if err != nil {
		return p, err
	}
	fmt.Fprint(o.Out, Outcome(p, append(left, checked...)))
	return p, nil
}

// readWorkspace reads the plugins' manifest claims and the names and
// folders the workspace's projects take.
func readWorkspace(root, version string, log io.Writer) ([]adoptmove.Claim, map[string]bool, error) {
	s, err := session.Open(session.Options{Dir: root, Version: version, Log: log})
	if err != nil {
		return nil, nil, err
	}
	defer s.Close()
	infos := make([]protocol.InitializeResult, len(s.Plugins.Clients))
	for i, c := range s.Plugins.Clients {
		infos[i] = c.Info
	}
	taken := map[string]bool{}
	for _, p := range s.Graph.Projects {
		taken[p.Name] = true
	}
	return adoptmove.ClaimsOf(infos), taken, nil
}

// settle identifies the project and settles its name, kind, folder and
// history. It refuses what the import cannot do, before anything changes.
func settle(root string, src *Source, claims []adoptmove.Claim, taken map[string]bool, o Options) (Plan, error) {
	c, err := adoptmove.Detect(src.Dir, claims)
	write := ""
	if errors.Is(err, adoptmove.ErrNoRootProject) {
		c, write, err = offer(src, claims, o)
	}
	if err != nil {
		return Plan{}, err
	}
	p := Plan{Name: c.Name, Ecosystem: c.Ecosystem, Manifest: c.Manifest, Source: *src, WriteManifest: write}
	if write != "" {
		for _, f := range manifestoffer.Sources(src.Dir, claims) {
			if f.Language == write {
				p.Sources = len(f.Files)
			}
		}
	}
	if o.Name != "" {
		p.Name = o.Name
	}
	if err := scaffold.ValidateName("project", p.Name); err != nil {
		return p, problem.New(err.Error(), "give another name with --name")
	}
	if taken[p.Name] {
		return p, problem.New("the workspace has a project named "+p.Name+" already", "give the import another name with --name")
	}
	var kind protocol.ProjectKind
	if o.As != "" {
		if kind, err = ParseKind(o.As); err != nil {
			return p, err
		}
	}
	switch {
	case o.Into != "" || kind != "":
		p.Folder, kind, err = Folder(kind, p.Name, o.Into)
	case o.DryRun:
		// A dry run changes nothing: it shows the suggested folder, and the
		// real run asks.
		p.Folder, kind, err = Folder(DefaultKind(c), p.Name, "")
	case operationplan.ShouldAsk(o.Env, o.Yes):
		p.Folder, kind, err = AskFolder(p.Name, DefaultKind(c), o.In, o.Out)
	default:
		suggested := DefaultKind(c)
		return p, problem.New(fmt.Sprintf("smith does not choose the folder of %s for you, and no terminal can answer (it looks like a %s)", p.Name, suggested),
			fmt.Sprintf("add --as %s (it goes into %s), or --into <folder>", suggested, mustFolder(suggested, p.Name)))
	}
	if err != nil {
		return p, err
	}
	p.Kind = string(kind)
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(p.Folder))); err == nil {
		return p, problem.New(p.Folder+" exists already, and smith import never writes into a folder that exists", "choose another folder with --into, or another name with --name")
	}
	if p.History, err = History(o.History, src.Git); err != nil {
		return p, err
	}
	p.Lockfiles = LockfilesIn(src.Dir, Lockfiles(p.Ecosystem))
	if p.History != HistoryCopy && !o.DryRun {
		if err := cleanTree(root); err != nil {
			return p, err
		}
	}
	return p, nil
}

// writeManifest has the language plugin write the manifest of the default
// stack into the placed folder, and records its file in the plan.
func writeManifest(root, version string, p *Plan, out io.Writer) error {
	s, err := session.Open(session.Options{Dir: root, Version: version, Log: io.Discard})
	if err != nil {
		return err
	}
	defer s.Close()
	w, err := manifestoffer.Write(s, p.WriteManifest, protocol.ProjectKind(p.Kind), p.Name, p.Folder)
	if err != nil {
		return err
	}
	stack := "the " + w.Plugin + " plugin's own"
	if w.Stack != "" {
		stack = "stack " + w.Stack
	}
	fmt.Fprintf(out, "  manifest: wrote %s (%s); the sources stay where they are\n", strings.Join(w.Files, ", "), stack)
	if len(w.Files) > 0 {
		p.Manifest = filepath.Base(filepath.FromSlash(w.Files[0]))
	}
	return nil
}

// LayoutLeft names what a written manifest leaves to the person: a Node
// template points its tsconfig and package.json at src/, and the sources of
// a folder without src/ stay at its top.
//
// Parameters: root is the workspace root; p is the plan.
// Returns: the item left, or nil.
func LayoutLeft(root string, p Plan) []Left {
	if p.WriteManifest != "node" {
		return nil
	}
	if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(p.Folder), "src")); err == nil && info.IsDir() {
		return nil
	}
	return []Left{{
		Problem: p.Name + ": the written package.json and tsconfig.json point at src/, and the sources are at the top of " + p.Folder,
		Remedy:  "move the sources into " + p.Folder + "/src (git mv), or point include, rootDir, main and exports at the folder's top\nsmith run typecheck build -p " + p.Name + " shows what still fails",
	}}
}

// cleanTree refuses the import of a git history on a work tree with uncommitted changes,
// which git itself refuses.
func cleanTree(root string) error {
	git := gitclient.New(root)
	if !git.HasCommits() {
		return problem.New("the workspace's repository has no commit yet, and the import of a git history needs one", "commit the workspace first, or use --history copy")
	}
	status, err := git.Run("status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) != "" {
		return problem.New("the work tree has uncommitted changes, and the import of a git history needs a clean one, because git makes commits",
			"commit or stash your changes, then run smith import again", "or use --history copy to copy the files without the history")
	}
	return nil
}

func mustFolder(kind protocol.ProjectKind, name string) string {
	dir, _, err := Folder(kind, name, "")
	if err != nil {
		return "a folder of its own"
	}
	return dir
}
