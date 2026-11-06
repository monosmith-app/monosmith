package doctorview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// onHeader moves the cursor to the heading of group gi.
func onHeader(t *testing.T, m Model, gi int) Model {
	t.Helper()
	for at, r := range m.layout() {
		if r.kind == headerRow && r.group == gi {
			m.Cursor = at
			return m
		}
	}
	t.Fatalf("no heading for group %d", gi)
	return m
}

func TestEnterOnAHeadingFoldsAndUnfoldsItsGroup(t *testing.T) {
	m := onHeader(t, sample(), 1) // web
	if !strings.Contains(plain(m), "Docker registries of web agree") {
		t.Fatal("groups start open")
	}
	m = m.Update(key("enter"))
	if out := plain(m); strings.Contains(out, "Docker registries of web agree") || !strings.Contains(out, "Go is installed") {
		t.Fatalf("only web's group folds:\n%s", out)
	}
	if !m.OnHeader() {
		t.Fatal("the cursor stays on the heading")
	}
	if m = m.Update(key("enter")); !strings.Contains(plain(m), "Docker registries of web agree") {
		t.Fatal("enter again unfolds")
	}
}

func TestETogglesEveryGroupAndTheHeadlineButtonSaysWhichWay(t *testing.T) {
	m := sample()
	if !strings.Contains(plain(m), "collapse all") {
		t.Fatalf("with every group open the button collapses:\n%s", plain(m))
	}
	m = m.Update(key("E"))
	out := plain(m)
	if strings.Contains(out, "Go is installed") || !strings.Contains(out, "expand all") {
		t.Fatalf("E folds every group and the button offers to expand:\n%s", out)
	}
	if out := plain(m.Update(key("E"))); !strings.Contains(out, "Go is installed") || !strings.Contains(out, "Docker registries of web agree") {
		t.Fatalf("E again opens them all:\n%s", out)
	}
}

func TestClickingTheHeadlineOrAHeadingFolds(t *testing.T) {
	m := sample().SelectRow(ToggleAllRow)
	if strings.Contains(plain(m), "Go is installed") {
		t.Fatal("a click on the headline folds every group")
	}
	m = sample().SelectRow(HeaderCode(0))
	if out := plain(m); strings.Contains(out, "Go is installed") || !strings.Contains(out, "Docker registries of web agree") {
		t.Fatalf("a click on a heading folds that group:\n%s", out)
	}
}

func TestAFoldedGroupKeepsItsCountsInTheHeading(t *testing.T) {
	m := onHeader(t, sample(), 1).Update(key("enter"))
	if !strings.Contains(plain(m), "web  1 failed") {
		t.Fatalf("the counts stay on the folded heading:\n%s", plain(m))
	}
}

func TestAFindingIsFiledUnderTheProjectItNamesNotTheOneItsTextMentions(t *testing.T) {
	findings := []doctor.Finding{
		{Finding: protocol.Finding{Check: "something about api", Project: "web", Detail: "mentions api"}},
		{Finding: protocol.Finding{Check: "git is installed", Area: "git"}},
		{Finding: protocol.Finding{Check: "node is installed", Area: "toolchain"}},
	}
	projects := []ProjectRef{{Name: "web", Root: "apps/web"}, {Name: "api", Root: "apps/api"}}
	got := GroupByProject(findings, projects)
	if len(got) != 2 || got[0].Project != "" || got[1].Project != "web" {
		t.Fatalf("groups: %+v", got)
	}
	// The workspace group lists the tools before git.
	if got[0].Indexes[0] != 2 || got[0].Indexes[1] != 1 {
		t.Fatalf("areas are ordered toolchain first: %+v", got[0])
	}
}

func TestAProjectTheListDoesNotKnowStillGetsItsGroup(t *testing.T) {
	findings := []doctor.Finding{{Finding: protocol.Finding{Check: "x", Project: "ghost"}}}
	got := GroupByProject(findings, []ProjectRef{{Name: "web"}})
	if len(got) != 1 || got[0].Project != "ghost" {
		t.Fatalf("groups: %+v", got)
	}
}

func TestTheCursorStartsOnTheFirstFindingNotTheHeading(t *testing.T) {
	if f, ok := sample().Selected(); !ok || f.Check != "Go is installed" {
		t.Fatalf("selected %q ok=%v", f.Check, ok)
	}
}
