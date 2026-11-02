package doctor

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestDefaultsVintageFinding(t *testing.T) {
	latest := workspace.LatestDefaultsVintage()
	if f := defaultsVintageFinding(latest, ""); !f.OK {
		t.Errorf("latest is not ok: %+v", f)
	}
	if f := defaultsVintageFinding("", ""); f.OK || !f.Warning || !strings.Contains(f.Remedy, "smith upgrade --defaults") {
		t.Errorf("none recorded: %+v", f)
	}
	if f := defaultsVintageFinding("2999.01", ""); f.OK || f.Remedy != "upgrade smith" {
		t.Errorf("newer: %+v", f)
	}
}

func TestDefaultsVintageFindingOffersToRecordTheCurrentDefaults(t *testing.T) {
	latest := workspace.LatestDefaultsVintage()
	f := defaultsVintageFinding("", "[workspace]\nname = \"x\"\n")
	if !strings.Contains(f.Detail, "keeps them exactly as they are") || f.Fix == nil || len(f.Fix.Choices) != 1 {
		t.Fatalf("%+v", f)
	}
	opts := f.Fix.Choices[0].Options
	if len(opts) != 2 || !strings.HasPrefix(opts[0].Label, "keep the current defaults") || !strings.HasPrefix(opts[1].Label, "move to the latest") {
		t.Fatalf("options: %+v", opts)
	}
	got := opts[0].Fix.Files[0]
	if got.Path != "smith.toml" || !strings.Contains(got.Content, `defaults = "`+latest+`"`) || !strings.Contains(got.Content, `name = "x"`) {
		t.Errorf("record option writes %+v", got)
	}
	if args := opts[1].Fix.Steps[0].Argv; !slices.Contains(args, "--defaults") {
		t.Errorf("move option runs %v", args)
	}
	if older := defaultsVintageFinding("2000.01", "[workspace]\n"); older.Fix == nil || len(older.Fix.Choices) != 0 {
		t.Errorf("an older recorded vintage has one remedy, the move: %+v", older.Fix)
	}
}