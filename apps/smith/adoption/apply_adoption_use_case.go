package adoption

import (
	"errors"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptdecision"
	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/operation"
	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ErrDirty is a step asked for with uncommitted changes.
var ErrDirty = errors.New("commit or stash your changes first: the adoption steps change the repository, and a clean tree keeps them reviewable")

// CreateBaselines tags, locally, the commits the lineage checks found under
// an old name with the project's current tag name; nothing is pushed.
func CreateBaselines(r Report, out io.Writer) (int, error) {
	git := gitclient.New(r.Root)
	existing, err := git.Tags("")
	if err != nil {
		return 0, err
	}
	made := 0
	for _, b := range r.Blockers {
		if b.Baseline == nil || slices.Contains(existing, b.Baseline.Name) {
			continue
		}
		if _, err := git.Run("tag", b.Baseline.Name, b.Baseline.Target); err != nil {
			return made, fmt.Errorf("tag %s: %w", b.Baseline.Name, err)
		}
		fmt.Fprintf(out, "  tagged %s (%s); push it with git push origin %s\n", b.Baseline.Name, strings.TrimSuffix(b.Baseline.Target, "^{commit}"), b.Baseline.Name)
		made++
	}
	return made, nil
}

// adoptionBlocked names each blocker of --init and the step that clears it, so
// the error says why (for an adopted workspace: that there is nothing to adopt).
func adoptionBlocked(r Report) error {
	var lines []string
	for _, b := range r.Blockers {
		if b.Blocks != BlocksAdoption {
			continue
		}
		line := b.Message
		if b.Step != "" {
			line += " (" + b.Step + ")"
		}
		lines = append(lines, line)
	}
	return errors.New("cannot adopt: " + strings.Join(lines, "; "))
}

// Init makes the repository a workspace: a pipeline file of the user's at a
// path smith owns is kept beside it as .pre-smith, smith.toml and the root files
// smith needs are written (theirs kept), the steps doctor would ask for (go.work,
// npm workspaces, and with install npm install) are taken, the pipelines are
// generated for the projects found. The choices c (line endings, agent files,
// CI providers, what happens to a pipeline of yours) are settled before, by
// adoptdecision.Resolve, and shown in the plan. Nothing is committed unless commit is true
// (smith adopt --commit) or the workspace sets commit_operations: the changes
// stay in the work tree, where git can take them back. It ends by saying what
// smith doctor still reports.
func Init(r Report, version string, install, eslint, commit bool, c adoptdecision.Choices, out io.Writer) error {
	if !r.Adoptable() {
		return adoptionBlocked(r)
	}
	if !r.Clean {
		return ErrDirty
	}
	git := gitclient.New(r.Root)
	own := OwnPipelines(r.Root)
	providers := adoptdecision.WithoutKept(c, own)
	moved := []string{cipipeline.DevcontainerPath}
	for _, p := range providers {
		moved = append(moved, ProviderPaths[p])
	}
	if len(providers) == 0 {
		moved = nil
	}
	for _, path := range moved {
		full := filepath.Join(r.Root, filepath.FromSlash(path))
		body, err := os.ReadFile(full)
		if err != nil || strings.Contains(string(body), GeneratedMarker) {
			continue
		}
		if _, err := git.Run("mv", path, path+".pre-smith"); err != nil {
			return fmt.Errorf("keep %s: %w", path, err)
		}
		fmt.Fprintf(out, "  kept your %s as %s.pre-smith\n", path, path)
	}
	if _, err := scaffold.NewWorkspace(scaffold.NewOptions{Name: r.Name, Dir: r.Root, Scope: r.Scope, Providers: providers, NoCI: len(providers) == 0, NoDevcontainer: len(providers) == 0,
		NoGitattributes: c.Gitattributes != adoptdecision.GitattributesAdd, SkipAgent: c.Agent == adoptdecision.AgentSkip, NoAgent: c.Agent == adoptdecision.AgentNever,
		Version: version, Out: out, NoCommit: true}); err != nil {
		return err
	}
	if len(r.Declared) > 0 {
		f, err := os.OpenFile(filepath.Join(r.Root, workspace.FileName), os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		_, werr := f.WriteString(declarationsText(r.Declared))
		if cerr := f.Close(); werr != nil || cerr != nil {
			return errors.Join(werr, cerr)
		}
		fmt.Fprintf(out, "  declared %d Go project(s) of the root module in %s\n", len(r.Declared), workspace.FileName)
	}
	if err := CompleteWorkspace(r, install, eslint, out); err != nil {
		return err
	}
	s, err := session.Open(session.Options{Dir: r.Root, Version: version, Log: out})
	if err != nil {
		return err
	}
	defer s.Close()
	contributions, err := cipipeline.Collect(s)
	if err != nil {
		return err
	}
	if _, _, err := cipipeline.Write(s.Root, s.Config, cipipeline.NewModel(contributions, cipipeline.SmithVersionFor(s.Config, version), s.Config.Release.Branch)); err != nil {
		return err
	}
	doctor.ReportCommandBugs(s, "smith adopt", out)
	if !commit && !s.Config.Workspace.CommitOperations {
		fmt.Fprintf(out, "  not committed: git status lists the changes and git can take them back; commit them when you are happy, or %s next time\n", operation.CommitNextTime("run smith adopt --commit"))
		return DescribeOutcome(r.Root, version, out)
	}
	if pending := operation.Always(r.Root); pending == nil {
		fmt.Fprintln(out, "  not committed: this folder is not a git work tree")
	} else if hash, err := pending.Finish("adopt smith"); err != nil {
		fmt.Fprintf(out, "  not committed (%v); commit when ready\n", strings.SplitN(err.Error(), "\n", 2)[0])
	} else if hash != "" {
		fmt.Fprintln(out, "  committed: chore: adopt smith (smith undo reverts it)")
	}
	return DescribeOutcome(r.Root, version, out)
}
