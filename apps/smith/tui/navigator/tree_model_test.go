package navigator

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func sample() Tree {
	return New([]Project{
		{Name: "web", Kind: "app", Language: "node", Tasks: []Task{{Name: "build"}, {Name: "test"}}, Deps: []string{"kit"}, Affected: true},
		{Name: "orders", Kind: "app", Language: "go", Tasks: []Task{{Name: "build"}}},
		{Name: "kit", Kind: "lib", Language: "go", Tasks: []Task{{Name: "test"}}, Dependents: []string{"web"}},
		{Name: "sdk", Kind: "package", Language: "python"},
	}).SetHeight(8)
}

func labels(t Tree) []string {
	var out []string
	for _, r := range t.Rows() {
		switch r.Kind {
		case GroupRow:
			out = append(out, "G:"+r.Group)
		case ProjectRow:
			out = append(out, "P:"+r.Project)
		case TaskRow:
			out = append(out, "T:"+r.Project+"/"+r.Task)
		}
	}
	return out
}

func TestALongProjectNameEndsInAnEllipsisAndDropsTheLanguageFirst(t *testing.T) {
	tree := New([]Project{
		{Name: "ms.sharepoint", Kind: "package", Language: "node"},
		{Name: "atlassian.confluence-client", Kind: "package", Language: "node"},
	}).SetHeight(8)
	out := strings.Join(tree.View(theme.Default(), 24, -1, true, nil), "\n")
	if !strings.Contains(out, "ms.sharepoint") || strings.Contains(out, "sharepoint node") {
		t.Fatalf("the language should give way before the name:\n%s", out)
	}
	if !strings.Contains(out, "…") || strings.Contains(out, "confluence-client") {
		t.Fatalf("a name that does not fit should end in an ellipsis:\n%s", out)
	}
}

func TestProjectsAreGroupedByKindAndSorted(t *testing.T) {
	got := strings.Join(labels(sample()), " ")
	if got != "G:app P:orders P:web G:lib P:kit G:package P:sdk" {
		t.Fatal(got)
	}
}

func TestArrowsExpandAProjectToItsTasksAndCollapseBack(t *testing.T) {
	tr := sample().Move(2).Expand() // web
	if got := strings.Join(labels(tr), " "); !strings.Contains(got, "P:web T:web/build T:web/test") {
		t.Fatal(got)
	}
	tr = tr.Move(1)
	if s := tr.Selected(); s.Kind != TaskRow || s.Task != "build" || s.Project != "web" {
		t.Fatalf("%+v", s)
	}
	tr = tr.Collapse() // a task row collapses to its parent
	if s := tr.Selected(); s.Kind != ProjectRow || s.Project != "web" {
		t.Fatalf("%+v", s)
	}
	tr = tr.Collapse()
	if got := strings.Join(labels(tr), " "); strings.Contains(got, "T:web") {
		t.Fatal(got)
	}
}

func TestAGroupCanBeClosed(t *testing.T) {
	tr := sample().Toggle()
	if got := strings.Join(labels(tr), " "); got != "G:app G:lib P:kit G:package P:sdk" {
		t.Fatal(got)
	}
}

func TestTheCursorStaysInRangeAndScrollsIntoView(t *testing.T) {
	tr := sample().SetHeight(3).End()
	if tr.Selected().Project != "sdk" || tr.top == 0 {
		t.Fatalf("%+v top %d", tr.Selected(), tr.top)
	}
	if tr.Move(5).Cursor() != tr.Cursor() {
		t.Fatal("moves past the end are clamped")
	}
	if sample().Move(-5).Cursor() != 0 {
		t.Fatal("moves before the start are clamped")
	}
}

func TestRowAtFindsTheRowUnderALineOfTheArea(t *testing.T) {
	tr := sample()
	if tr.RowAt(0) != 0 || tr.RowAt(1) != 1 || tr.RowAt(7) != -1 || tr.RowAt(-1) != -1 {
		t.Fatalf("%d %d %d", tr.RowAt(0), tr.RowAt(1), tr.RowAt(7))
	}
	tr = tr.SetHeight(3).Scroll(2)
	if tr.RowAt(0) != 2 {
		t.Fatalf("scrolled: %d", tr.RowAt(0))
	}
}

func TestClickingARowSelectsIt(t *testing.T) {
	tr := sample().Select(sample().RowAt(3))
	if tr.Selected().Group != "lib" {
		t.Fatalf("%+v", tr.Selected())
	}
}

func TestSelectProjectOpensItsGroup(t *testing.T) {
	tr := sample().Toggle().SelectProject("orders")
	if tr.Selected().Project != "orders" {
		t.Fatalf("%+v", tr.Selected())
	}
}

func TestViewShowsAffectedDotsLanguagesAndAStatusSuffix(t *testing.T) {
	lines := sample().View(theme.Default(), 30, -1, true, func(p, task string) string {
		if p == "web" && task == "" {
			return " ✖"
		}
		return ""
	})
	text := strings.Join(lines, "\n")
	for _, want := range []string{"▾ Apps", "● web", "○ orders", "go", "node", "✖", "Libs", "Packages"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in\n%s", want, text)
		}
	}
	for _, l := range lines {
		if len([]rune(l)) < 30 && !strings.Contains(l, "\x1b") {
			t.Fatalf("line not padded: %q", l)
		}
	}
}

func TestAnEmptyTreeSaysHowToAddAProject(t *testing.T) {
	text := strings.Join(New(nil).View(theme.Default(), 30, -1, true, nil), "\n")
	if !strings.Contains(text, "No projects") {
		t.Fatal(text)
	}
	if New(nil).Selected().Valid {
		t.Fatal("nothing to select")
	}
	_ = New(nil).Move(3).Toggle().Collapse().Expand().Page(1).Home().End()
}
