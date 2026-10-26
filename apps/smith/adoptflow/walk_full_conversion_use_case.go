package adoptflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/fullconversion"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ErrNotConvertible is a conversion step asked of a repository smith adopts
// instead of converting.
var ErrNotConvertible = errors.New("this repository is not an Nx or Turborepo workspace, so there is nothing to convert")

// Preview is the plan of a full conversion for a set of choices: the data,
// and the review the wizard shows (every group as diffs, what stays and why,
// the closing checks).
type Preview struct {
	Report fullconversion.Report
	Text   string
	// Changes are the files the conversion itself writes, for the summary above
	// the review (Text lists them, then the removals).
	Changes Changes
}

// Prepare converts the workspace on a scratch copy, so what the old tooling
// would be after the conversion can be worked out without touching the real
// files, and returns the choices the workspace's own tooling suggests with the
// questions that ask them.
func (s *Session) Prepare() (convertchoices.Choices, []convertchoices.Question, error) {
	if s.analysis == nil {
		return convertchoices.Choices{}, nil, ErrNotConvertible
	}
	if s.view == nil {
		view, err := s.analysis.MakeView()
		if err != nil {
			return convertchoices.Choices{}, nil, fmt.Errorf("could not work out what the conversion leaves behind: %w", err)
		}
		s.view = view
	}
	usage := legacyinventory.Detect(s.view.Tree)
	def := convertchoices.Defaults(usage)
	return def, convertchoices.Questions(usage, def), nil
}

// PlanFull plans the removals the choices ask for on the converted scratch
// copy. verify adds the build, test, lint and doctor checks to the closing
// steps.
func (s *Session) PlanFull(c convertchoices.Choices, verify bool) (Preview, error) {
	if s.view == nil {
		if _, _, err := s.Prepare(); err != nil {
			return Preview{}, err
		}
	}
	report := fullconversion.MakeReport(s.view.Tree, c, verify)
	report.Finish = removalplan.Finish(convert.SmithCommand(), verify)
	var b strings.Builder
	changes, err := s.Changes()
	if err == nil {
		b.WriteString("● The files, in detail\n")
		b.WriteString(changes.Details())
		b.WriteString("\n")
	}
	if err := fullconversion.Describe(&b, s.view.Dir, report, true); err != nil {
		return Preview{}, err
	}
	return Preview{Report: report, Text: b.String(), Changes: changes}, nil
}

// Convert applies the conversion to the repository, without installing: the
// closing steps do that. It writes "… " and a name to out as the one step
// starts, and returns the summary of what changed.
func (s *Session) Convert(_ context.Context, c convertchoices.Choices, out io.Writer) (string, error) {
	if s.analysis == nil {
		return "", ErrNotConvertible
	}
	fmt.Fprintln(out, "… "+ConvertStepName)
	log := io.Discard
	return s.analysis.Convert(c.Level == convertchoices.LevelFull && c.CI == convertchoices.Keep, log)
}

// Result is what the conversion found worth telling, once it was applied.
func (s *Session) Result() convert.Highlights {
	if s.analysis == nil {
		return convert.Highlights{}
	}
	return convert.HighlightsOf(s.analysis.Plan)
}

// ConvertStepName is the one step Convert reports.
const ConvertStepName = "convert the workspace to smith"

// NextGroup plans the removals again from the files as they are now and
// returns the plan: its first group is the next to ask about. A group in
// declined is left out, and the groups after it are planned without it.
func (s *Session) NextGroup(c convertchoices.Choices, declined map[string]bool) (removalplan.Plan, error) {
	files, err := s.git.Files()
	if err != nil {
		return removalplan.Plan{}, err
	}
	return removalplan.Build(legacyinventory.NewTree(s.Root, files), c, declined), nil
}

// ShowFix writes a planned fix as the review shows it, against the
// repository's files as they are now.
func (s *Session) ShowFix(p doctorfix.Planned) (string, error) {
	var b strings.Builder
	err := doctorfix.ShowFix(&b, s.Root, p)
	return b.String(), err
}

// ApplyFix writes a fix's files and runs its steps from the repository root,
// stopping at the first that fails (a *doctorfix.StepError). It writes a line
// "… name" to out as each step starts.
func (s *Session) ApplyFix(ctx context.Context, fix protocol.Fix, out io.Writer) error {
	return doctorfix.Apply(ctx, s.Root, fix, out)
}

// FinishFix is the last stage of a conversion: one install, the lockfile
// check, then the proof that it works: the build, the tests and doctor, and
// the lint too when withLint (the full level, where smith owns the lint).
func (s *Session) FinishFix(withLint bool) doctorfix.Planned {
	if withLint {
		return removalplan.Finish(convert.SmithCommand(), true)
	}
	return removalplan.FinishBuildAndTest(convert.SmithCommand())
}

// Report is the full report of the applied conversion, as it was saved.
func (s *Session) Report() (string, error) {
	raw, err := os.ReadFile(filepath.Join(s.Root, filepath.FromSlash(convert.ReportFile)))
	return string(raw), err
}

// AdoptFix is the conservative adoption of a plain repository as a fix: one
// step that runs smith adopt --init, which writes smith.toml and the
// pipelines and commits.
func (s *Session) AdoptFix() doctorfix.Planned {
	return doctorfix.Planned{
		Fix: protocol.Fix{
			Description: "Make this repository a smith workspace: write smith.toml and the pipelines, and commit them",
			Steps: []protocol.FixStep{{
				Description: "adopt the repository (smith.toml, the CI pipelines, one commit)",
				Argv:        []string{convert.SmithCommand(), "adopt", "--init"},
			}},
		},
		Checks: []string{"adopt"},
	}
}
