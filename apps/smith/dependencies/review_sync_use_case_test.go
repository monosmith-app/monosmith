package dependencies

import (
	"bufio"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func axiosDrift() Drift {
	site := func(project, spec string) protocol.Declaration {
		return protocol.Declaration{Project: project, File: project + "/package.json", Name: "axios", Spec: spec, Section: "dependencies", Rewritable: true}
	}
	sites := []protocol.Declaration{site("api", "^1.2.0"), site("web", "^1.6.0"), site("cli", "^1.4.0")}
	return classify(Drift{Plugin: "node", Name: "axios", Reason: "installed", Sites: sites, Installed: "1.6.0"}, "^1.6.0")
}

func settle(t *testing.T, r Reviewer, input string) ([]Drift, string, error) {
	t.Helper()
	var out strings.Builder
	r.Out = &out
	r.In = bufio.NewReader(strings.NewReader(input))
	got, err := r.Settle([]Drift{axiosDrift()})
	return got, out.String(), err
}

func TestCandidatesListEveryWayToConverge(t *testing.T) {
	var specs []string
	for _, c := range Candidates(axiosDrift()) {
		specs = append(specs, c.Spec)
	}
	if want := []string{"^1.6.0", "^1.2.0", "^1.4.0"}; !slices.Equal(specs, want) {
		t.Fatalf("got %v, want %v", specs, want)
	}
}

func TestPreferRules(t *testing.T) {
	for _, tc := range []struct{ rule, want string }{
		{PreferHighest, "^1.6.0"},
		{PreferLowest, "^1.2.0"},
		{PreferInstalled, "^1.6.0"},
		{PreferPick, "^1.6.0"},
	} {
		got, ok := Prefer(axiosDrift(), tc.rule)
		if !ok || got.Target != tc.want {
			t.Errorf("%s: got %q (ok %v), want %q", tc.rule, got.Target, ok, tc.want)
		}
	}
	noInstalled := axiosDrift()
	noInstalled.Installed = ""
	if _, ok := Prefer(noInstalled, PreferInstalled); ok {
		t.Error("installed should not apply when nothing is installed")
	}
	low, _ := Prefer(axiosDrift(), PreferLowest)
	if len(low.Fixable) != 2 {
		t.Errorf("the lowest version rewrites the two other projects, got %d", len(low.Fixable))
	}
}

func TestSettleRefusesWithoutATerminalAndNamesTheFlags(t *testing.T) {
	_, out, err := settle(t, Reviewer{}, "")
	if err == nil {
		t.Fatal("want a refusal")
	}
	for _, want := range []string{"--yes", "--prefer", "--to", "--dry-run"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal lacks %s:\n%v", want, err)
		}
	}
	if !strings.Contains(out, "api") || !strings.Contains(out, "^1.2.0 -> ^1.6.0") {
		t.Errorf("the plan comes before the refusal:\n%s", out)
	}
}

func TestSettleYesWritesThePlanAsItStandsAndSaysHowToChangeIt(t *testing.T) {
	got, out, err := settle(t, Reviewer{Yes: true}, "")
	if err != nil || got[0].Target != "^1.6.0" {
		t.Fatalf("%v %v", got, err)
	}
	if !strings.Contains(out, "--prefer") {
		t.Errorf("the output should name the flag that changes the version:\n%s", out)
	}
}

func TestSettleFlagsPickTheWinner(t *testing.T) {
	got, _, err := settle(t, Reviewer{Yes: true, Prefer: PreferLowest}, "")
	if err != nil || got[0].Target != "^1.2.0" {
		t.Fatalf("--prefer lowest: %v %v", got[0].Target, err)
	}
	got, _, err = settle(t, Reviewer{Yes: true, Prefer: PreferHighest, To: map[string]string{"axios": "^1.4.0"}}, "")
	if err != nil || got[0].Target != "^1.4.0" {
		t.Fatalf("--to wins over --prefer: %v %v", got[0].Target, err)
	}
}

func TestSettleDryRunAsksNothingAndChangesNothing(t *testing.T) {
	_, out, err := settle(t, Reviewer{Preview: true}, "")
	if err != nil || !strings.Contains(out, "Dry run: nothing was changed.") {
		t.Fatalf("%v\n%s", err, out)
	}
}

func TestSettleInATerminal(t *testing.T) {
	asks := Reviewer{Asks: true}
	for _, tc := range []struct {
		name, input, target string
		err                 error
	}{
		{"enter then yes keeps smith's pick", "\ny\n", "^1.6.0", nil},
		{"the lowest then yes", "3\ny\n", "^1.2.0", nil},
		{"one by one picks a version", "4\n3\ny\n", "^1.4.0", nil},
		{"cancel", "5\n", "", ErrNotConfirmed},
		{"no at the end", "\nn\n", "", ErrNotConfirmed},
		{"a closed input does not write", "\n", "", ErrNotConfirmed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, out, err := settle(t, asks, tc.input)
			if err != tc.err {
				t.Fatalf("err %v, want %v\n%s", err, tc.err, out)
			}
			if tc.err == nil && got[0].Target != tc.target {
				t.Fatalf("target %s, want %s\n%s", got[0].Target, tc.target, out)
			}
		})
	}
}

func TestMajorsAreSeparatedFromTheRest(t *testing.T) {
	us := []Update{{Name: "a", Level: "patch"}, {Name: "b", Level: "major"}, {Name: "c", Level: "minor"}}
	if got := Majors(us); len(got) != 1 || got[0].Name != "b" {
		t.Errorf("Majors: %v", got)
	}
	if got := WithoutMajors(us); len(got) != 2 || got[0].Name != "a" || got[1].Name != "c" {
		t.Errorf("WithoutMajors: %v", got)
	}
}
