package adoptflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoption"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/fullconversion"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// IssueFullForPlain is the issue that tracks the full level for repositories
// that are not Nx or Turborepo workspaces.
const IssueFullForPlain = adoption.FullLevelIssue

// NoFullForPlain says why a plain repository has no full level yet.
const NoFullForPlain = "The full level (smith becomes the only tool) applies to Nx and Turborepo workspaces, which smith converts. A plain repository adopts at the conservative level for now: it keeps its build and lint setup. Tracked in " + IssueFullForPlain + "."

// Session is one repository being adopted: what it is, and the steps that
// bring it under smith. Open reads it; the other calls each do one step.
type Session struct {
	// Root is the repository's absolute path.
	Root    string
	Version string

	found    Found
	analysis *convert.Analysis
	report   adoption.Report
	git      *gitclient.Git
	view     *fullconversion.View
}

// Open reads the repository at dir and says what it is. It changes nothing;
// whatever stops the wizard is in Found.Blockers, never an error, so the
// wizard can tell it in its first screen. An error is a repository that
// could not be read at all.
func Open(dir, version string) (*Session, error) {
	return OpenWithProgress(dir, version, nil)
}

// OpenWithProgress is Open that tells progress each stage of the read, in the
// words convert uses ("Asking Nx to describe the projects…"). A nil progress
// is ignored.
func OpenWithProgress(dir, version string, progress convert.Progress) (*Session, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	s := &Session{Root: root, Version: version, git: gitclient.New(root)}
	s.found = Found{Name: filepath.Base(root), Source: "repository"}
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
		return err == nil
	}
	switch {
	case exists(workspace.FileName):
		s.found.Kind = KindSmith
		s.found.Blockers = []string{"This is already a smith workspace (" + workspace.FileName + " exists). smith doctor checks it; smith upgrade brings it up to this smith."}
		return s, nil
	case !s.git.IsRepository():
		s.found.Kind = KindNotARepo
		s.found.Blockers = []string{`This folder is not a git repository: smith makes its changes through git, so each one can be reviewed and undone. Run git init && git add -A && git commit -m "before smith", then open this again.`}
		return s, nil
	case exists("nx.json") || exists("turbo.json") || exists("turbo.jsonc"):
		if err := s.openConvertible(exists, progress); err != nil {
			return nil, err
		}
		return s, nil
	}
	if err := s.openPlain(); err != nil {
		return nil, err
	}
	return s, nil
}

// Found is what the repository is.
func (s *Session) Found() Found { return s.found }

// Close removes the scratch copy of the converted workspace, if one was made.
func (s *Session) Close() {
	if s.view != nil {
		s.view.Close()
		s.view = nil
	}
}

// openConvertible reads an Nx or Turborepo workspace through convert.
func (s *Session) openConvertible(exists func(string) bool, progress convert.Progress) error {
	s.found.Kind = KindNx
	s.found.Source = "Nx workspace"
	if exists("turbo.json") || exists("turbo.jsonc") {
		s.found.Kind, s.found.Source = KindTurborepo, "Turborepo workspace"
	}
	a, err := convert.AnalyseWithProgress(s.Root, s.Version, progress)
	if err != nil {
		var p *problem.Error
		switch {
		case errors.As(err, &p):
			text := p.Problem
			if len(p.Remedy) > 0 {
				text += " → " + strings.Join(p.Remedy, "; or ")
			}
			s.found.Blockers = append(s.found.Blockers, text)
		case s.found.Kind == KindNx && (!exists("node_modules/nx") || strings.Contains(err.Error(), "nx graph failed")):
			// smith reads an Nx workspace's projects from Nx itself, which needs
			// the packages the repository lists to be the ones installed.
			s.found.NeedsInstall = true
			s.found.Blockers = append(s.found.Blockers, "smith needs to read your projects before it can move them, and Nx describes them, so the packages this repository lists have to be installed first. smith can do that for you (about 2 minutes).")
		default:
			s.found.Blockers = append(s.found.Blockers, err.Error())
		}
		return nil
	}
	s.analysis = a
	if progress != nil {
		progress("Found " + a.Counts + ".")
	}
	s.found.Source = a.Source
	s.found.Name = a.Plan.Config.Workspace.Name
	for _, p := range a.Plan.Projects {
		if p.Supported {
			s.found.Projects = append(s.found.Projects, Project{Name: p.Name, Root: p.Root, Language: p.Language, Kind: string(p.Kind)})
		}
	}
	s.found.Tooling = legacyinventory.ToolsInUse(legacyinventory.NewTree(s.Root, a.Files))
	if s.found.Kind == KindTurborepo {
		// Turborepo is the tool smith replaces; the inventory names the tools around it.
		s.found.Tooling = append([]string{"Turborepo"}, s.found.Tooling...)
	}
	s.found.Full = true
	if a.Blocked {
		s.found.Blockers = append(s.found.Blockers, strings.TrimSpace(a.Verdict)+" smith convert --dry-run lists them.")
	}
	if a.Dirty {
		s.found.Blockers = append(s.found.Blockers, `The working tree has uncommitted changes: the conversion would not be one reviewable change, and git is the way back. Run git add -A && git commit -m "before smith", then open this again.`)
	}
	return nil
}

// openPlain reads a repository that is not an Nx or Turborepo workspace
// through smith adopt.
func (s *Session) openPlain() error {
	s.found.Kind = KindPlain
	r, err := adoption.Inspect(s.Root, s.Version, io.Discard)
	if err != nil {
		return err
	}
	s.report = r
	s.found.Name = r.Name
	s.found.RootProject = r.RootProject
	for _, p := range r.Projects {
		s.found.Projects = append(s.found.Projects, Project{Name: p.Name, Root: p.Root, Language: p.Language, Kind: p.Kind})
	}
	for _, b := range r.Blockers {
		text := b.Message
		if b.Step != "" {
			text += " → " + b.Step
		}
		if b.Blocks == adoption.BlocksAdoption {
			s.found.Blockers = append(s.found.Blockers, text)
		} else {
			s.found.Warnings = append(s.found.Warnings, text+" (needs clearing before the first release)")
		}
	}
	for _, w := range r.Warnings {
		s.found.Warnings = append(s.found.Warnings, w.Message)
	}
	if !r.Clean && len(s.found.Blockers) == 0 {
		s.found.Blockers = append(s.found.Blockers, `The working tree has uncommitted changes: adopt makes one commit, and a clean tree keeps it reviewable. Run git add -A && git commit -m "before smith", then open this again.`)
	}
	if files, err := s.git.Files(); err == nil {
		s.found.Tooling = legacyinventory.ToolsInUse(legacyinventory.NewTree(s.Root, files))
	}
	s.found.NoFull = NoFullForPlain
	return nil
}

// Install installs the packages the repository lists, the way its lockfile
// says (npm ci when there is one), and reads the repository again.
func (s *Session) Install(ctx context.Context, out io.Writer) error {
	fmt.Fprintln(out, "… install the packages")
	return convert.InstallPackages(ctx, s.Root)
}
