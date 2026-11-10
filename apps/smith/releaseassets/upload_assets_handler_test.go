package releaseassets

import (
	"context"
	"io"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// The command takes the install step from the package that owns the CI
// phases, so `smith release assets --source tag` and `smith ci assets` prepare
// the worktree of a tag the same way.
func TestCommandTakesTheInstallStepOfATagWorktree(t *testing.T) {
	called := false
	prepare := func(io.Writer) func(context.Context, *session.Session) error {
		called = true
		return func(context.Context, *session.Session) error { return nil }
	}
	cmd := Command("test", prepare)
	if cmd.Use != "assets" || cmd.Flags().Lookup("source") == nil {
		t.Fatalf("the command lost its shape: %s", cmd.Use)
	}
	if called {
		t.Fatal("the install step is made when the command runs, not when it is built")
	}
}
