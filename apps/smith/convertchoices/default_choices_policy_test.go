package convertchoices

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
)

func TestDefaultsFollowWhatTheWorkspaceUses(t *testing.T) {
	got := Defaults(legacyinventory.Usage{ESLint: true, SharedLintConfig: "@mnci/eslint-config", Vitest: true})
	if got.Lint != Smith || got.Test != Vitest || got.Build != Smith || got.CI != Smith || got.Level != LevelFull {
		t.Fatalf("defaults = %+v", got)
	}
	// A config written by hand is not replaced unless asked.
	if got := Defaults(legacyinventory.Usage{ESLint: true, Jest: true}); got.Lint != Keep || got.Test != Jest {
		t.Fatalf("defaults = %+v", got)
	}
	if got := Defaults(legacyinventory.Usage{}); got.Test != Keep || got.Lint != Keep {
		t.Fatalf("defaults = %+v", got)
	}
}
