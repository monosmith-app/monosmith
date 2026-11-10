package sources

import (
	"context"

	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

// OutsideSources are the sources of a shell that is not in a workspace: it can
// show a plan (Capture) and run a command in the shell (Start), as smith new
// and smith convert need, but it loads no view, because there is none.
//
// Parameters: dir is the folder the shell was opened in.
// Returns: the sources, with only Capture, Open and Start set.
func OutsideSources(dir string) Sources {
	return Sources{
		Capture: func(ctx context.Context, argv []string) Captured { return CaptureCommand(ctx, dir, argv) },
		Open:    OpenPath,
		Start: func(argv []string) (runner.Handle, error) {
			p, err := runner.Start(context.Background(), runner.Spec{Argv: argv, Dir: dir})
			if err != nil {
				return nil, err
			}
			return p, nil
		},
	}
}
