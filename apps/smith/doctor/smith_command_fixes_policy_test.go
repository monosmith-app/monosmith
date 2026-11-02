package doctor

import (
	"slices"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// stepsOf lists every smith step a fix carries, in its choices too.
func stepsOf(f *protocol.Fix) [][]string {
	if f == nil {
		return nil
	}
	var out [][]string
	for _, s := range f.Steps {
		out = append(out, s.Argv)
	}
	for _, c := range f.Choices {
		for _, o := range c.Options {
			out = append(out, stepsOf(&o.Fix)...)
		}
	}
	return out
}

func TestEveryFindingWithASmithRemedyCarriesAFixThatRunsInTheInterface(t *testing.T) {
	behind := defaultsVintageFinding("2000.01", "")
	cache := localCacheFindings(cachemaintenance.Status{Exists: true, Bytes: 3 << 30, LimitBytes: 2 << 30, BrokenEntries: 4})
	lib, _ := UnusedLibHint("core", 0, 3, time.Now().AddDate(0, -3, 0), time.Now())
	pkg, _ := UnpublishedPackageHint("sdk", true)
	drift := deriveFindings(derive.Plan{Items: []derive.Item{
		{File: protocol.DerivedFile{Path: "a.json", Renderer: "r"}, Status: derive.StatusDrift},
		{File: protocol.DerivedFile{Path: "b.json", Renderer: "r"}, Status: derive.StatusStale},
	}})
	all := append([]protocol.Finding{behind, lib, pkg}, cache...)
	all = append(all, drift...)
	for _, f := range all {
		if f.OK {
			continue
		}
		if f.Fix == nil || f.Fix.Safe {
			t.Errorf("%q: a remedy that names a smith command needs an unsafe fix the interface can run, got %+v", f.Check, f.Fix)
			continue
		}
		steps := stepsOf(f.Fix)
		if len(steps) == 0 {
			t.Errorf("%q: the fix runs nothing", f.Check)
		}
		for _, argv := range steps {
			if argv[0] != "smith" {
				t.Errorf("%q: step %v does not run smith", f.Check, argv)
			} else if _, ok := catalog.Find(argv[1:]); !ok {
				t.Errorf("%q: step %v is no action of the catalogue", f.Check, argv)
			}
		}
	}
}

func TestADerivedFileThatWasEditedAsksWhoOwnsItAndEachAnswerRunsOneDeriveFlag(t *testing.T) {
	f := deriveFindings(derive.Plan{Items: []derive.Item{{File: protocol.DerivedFile{Path: "a.json", Renderer: "r"}, Status: derive.StatusDrift}}})[0]
	opts := f.Fix.Choices[0].Options
	if len(f.Fix.Choices) != 1 || len(opts) != 3 {
		t.Fatalf("%+v", f.Fix)
	}
	want := [][]string{
		{"smith", "derive", "--replace", "a.json", "--yes"},
		{"smith", "derive", "--adopt", "a.json", "--yes"},
		{"smith", "derive", "--eject", "a.json", "--yes"},
	}
	for i, o := range opts {
		if !slices.Equal(o.Fix.Steps[0].Argv, want[i]) {
			t.Errorf("option %d runs %v, want %v", i+1, o.Fix.Steps[0].Argv, want[i])
		}
	}
}
