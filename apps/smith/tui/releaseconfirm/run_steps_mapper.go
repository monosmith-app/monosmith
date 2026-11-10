package releaseconfirm

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Steps are the lines drawn while the release runs, one per package: waiting,
// building (with the spinner), tagged or stopped. The state comes from the
// lines the release printed ("lib: building 0.1.0", "lib: tagged <tag>").
//
// Parameters: items are the packages of the release; output is what it printed;
// tick is the animation frame; ended is true once the process ended and ok says
// how.
// Returns: the lines, in plan order.
func Steps(th theme.Theme, items []release.Item, output []string, tick int, ended, ok bool) []string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		state := stateOf(item, output)
		mark, text := th.Muted(th.Icons.Queued), th.Muted(words.Say("tui.release_run.queued"))
		switch {
		case state == tagged:
			mark, text = th.Success(th.Icons.Passed), words.Say("tui.release_run.tagged", "tag", item.Tag)
		case state == building && ended && !ok:
			mark, text = th.Failure(th.Icons.Failed), th.Failure(words.Say("tui.release_run.failed"))
		case state == building && !ended:
			mark, text = th.Info(th.Spin(tick)), words.Say("tui.release_run.building")
		}
		lines = append(lines, words.Say("tui.release_run.step", "mark", mark, "project", item.Project, "state", text))
	}
	return lines
}

type stepState int

const (
	waiting stepState = iota
	building
	tagged
)

// stateOf reads how far a package went from the output.
func stateOf(item release.Item, output []string) stepState {
	state := waiting
	for _, l := range output {
		l = pane.Strip(l)
		switch {
		case strings.HasPrefix(l, item.Project+": tagged "):
			return tagged
		case strings.HasPrefix(l, item.Project+":"):
			state = building
		}
	}
	return state
}
