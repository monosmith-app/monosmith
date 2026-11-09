package tui

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/toast"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// completionArgv is the command line that sets up tab completion once the
// person said yes.
var completionArgv = []string{"completion", "install", "--yes"}

// isCompletionSetup is true for the command line that writes the completion
// script.
func isCompletionSetup(argv []string) bool {
	return len(argv) > 1 && argv[0] == "completion" && argv[1] == "install"
}

// openCompletion asks smith which files and lines would be written (the dry
// run of the install) and shows them in the confirmation. Nothing is written
// until the person confirms; the confirmed line then runs in the run panel.
func (s *Shell) openCompletion() tea.Cmd {
	title, why := commands.Describe(completionArgv)
	return s.captureInDialog(capPlan, title, why, completionArgv, append(slices.Clone(completionArgv), "--dry-run"), words.Say("tui.completion.working_out_the_files"))
}

// declineCompletion remembers "Not now" for good and takes the item off Home.
func (s *Shell) declineCompletion() {
	if s.sources.DeclineCompletion != nil {
		s.sources.DeclineCompletion()
	}
	s.completionWanted = false
	s.notify(words.Say("tui.completion.declined"), toast.Info)
}
