package adoptwizard

import (
	"context"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Flow is what the wizard asks of smith; *adoptflow.Session is the real one.
type Flow interface {
	Found() adoptflow.Found
	// Prepare returns the choices the repository's own tooling suggests and
	// the questions that ask them.
	Prepare() (convertchoices.Choices, []convertchoices.Question, error)
	// PlanFull plans the full conversion for the choices.
	PlanFull(c convertchoices.Choices, verify bool) (adoptflow.Preview, error)
	// Convert applies the conversion, writing "… name" as its step starts, and
	// returns its summary.
	Convert(ctx context.Context, c convertchoices.Choices, out io.Writer) (string, error)
	// Changes lists the files the conversion creates, edits and removes; it
	// needs Prepare to have run.
	Changes() (adoptflow.Changes, error)
	// Result is what the applied conversion found worth telling.
	Result() convert.Highlights
	// Install installs the packages the repository lists.
	Install(ctx context.Context, out io.Writer) error
	// NextGroup plans the removals again from the files as they are.
	NextGroup(c convertchoices.Choices, declined map[string]bool) (removalplan.Plan, error)
	// ShowFix writes a fix as the review shows it.
	ShowFix(p doctorfix.Planned) (string, error)
	// ApplyFix applies a fix's files and steps, writing "… name" as each starts.
	ApplyFix(ctx context.Context, fix protocol.Fix, out io.Writer) error
	// FinishFix is the install and the checks that close a conversion: the
	// build, the tests and doctor, and the lint too with withLint.
	FinishFix(withLint bool) doctorfix.Planned
	// Report is the full report of the applied conversion, as it was saved.
	Report() (string, error)
	// AdoptFix is the conservative adoption of a plain repository.
	AdoptFix() doctorfix.Planned
	// Close releases what the flow holds (a scratch copy).
	Close()
}

// Opener reads the repository and returns its Flow. It may write "… name"
// lines to out to tell what it is reading; the dialog shows the latest.
type Opener func(io.Writer) (Flow, error)

// Job is a step that takes time: the shell runs it off the UI goroutine,
// reports each "… name" line it writes to out as Progress, and its end as Done.
type Job struct {
	// Name says what it does, for the status bar.
	Name string
	// Steps is how many "… name" lines it writes (0 when it writes none).
	Steps int
	Run   func(ctx context.Context, out io.Writer) (any, error)
}

// Outcome is what a key or a click asked of the shell.
type Outcome struct {
	// Close is true once the dialog should go away.
	Close bool
	// Note is what to tell when it closes ("" for the usual).
	Note string
	// Job is a step to run; nil for none.
	Job *Job
	// Stop is true when the user asked to cancel the job that is running.
	Stop bool
	// Copy is text for the clipboard ("" for none).
	Copy string
	// Finished is set when the repository is a workspace now: the shell opens
	// it, because it was started outside one.
	Finished string
	// Build asks the shell to start building everything once it is open.
	Build bool
}

// Stage is where the walk is.
type Stage int

// The stages, in the order a full conversion goes through them.
const (
	// StageReading reads the repository.
	StageReading Stage = iota
	// StageFound tells what was found.
	StageFound
	// StageLevel asks how far to go.
	StageLevel
	// StagePreparing converts a scratch copy to see what would be left.
	StagePreparing
	// StageInstalling installs the packages the repository lists, then reads it again.
	StageInstalling
	// StageListing works out the files a conservative conversion changes.
	StageListing
	// StageChoices asks the full level's choices.
	StageChoices
	// StagePlanning plans the removals for the choices.
	StagePlanning
	// StageReview shows the whole plan to be confirmed.
	StageReview
	// StageConverting applies the conversion (or the adoption).
	StageConverting
	// StageChecking plans the next group from the files as they are.
	StageChecking
	// StageGroup asks about one group of removals.
	StageGroup
	// StageRemoving applies the group that was agreed.
	StageRemoving
	// StageFinishing installs and checks the workspace.
	StageFinishing
	// StageDone tells what happened.
	StageDone
	// StageReport shows the saved report of the conversion; Enter goes back to
	// the result.
	StageReport
)
