package projectpage

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

var plain = theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode})

// sampleList is two apps, a lib and a package, with every kind of mark.
func sampleList() ListInput {
	return ListInput{
		Rows: []ListRow{
			{Index: 0, Name: "orders", Kind: "app", Language: "go", Mark: Failed, Changed: true},
			{Index: 1, Name: "web", Kind: "app", Language: "node", Mark: Passed},
			{Index: 2, Name: "kit", Kind: "lib", Language: "go", Mark: NotRun},
			{Index: 3, Name: "sdk", Kind: "package", Language: "python", Mark: Cached},
		},
		Kinds:   []string{"app", "lib", "package"},
		Totals:  map[string]int{"app": 2, "lib": 1, "package": 1},
		Changed: 1, Selected: 1,
	}
}

func TestTheProjectsAreGroupedUnderTheirKindWithACount(t *testing.T) {
	lines, rowOf, _ := ListBody(plain, sampleList(), 60)
	var text []string
	for _, l := range lines {
		text = append(text, strings.TrimRight(pane.Strip(l), " "))
	}
	want := []string{"APPS 2", "×  orders", "✓  web", "LIBS 1", "–  kit", "PACKAGES 1", "✓  sdk"}
	if len(text) != len(want) {
		t.Fatalf("got %d lines, want %d:\n%s", len(text), len(want), strings.Join(text, "\n"))
	}
	for i, w := range want {
		if !strings.Contains(text[i], w) {
			t.Fatalf("line %d is %q, want it to hold %q", i, text[i], w)
		}
	}
	for line, idx := range rowOf {
		if !strings.Contains(text[line], sampleList().Rows[idx].Name) {
			t.Fatalf("line %d (%q) answers a click with project %d", line, text[line], idx)
		}
	}
	if _, ok := rowOf[0]; ok {
		t.Fatal("a heading is not a project")
	}
}

func TestEachRowLeadsWithTheMarkOfItsLastRunAndTheChangedMarkInTheAccent(t *testing.T) {
	lines, _, _ := ListBody(plain, sampleList(), 60)
	orders := pane.Strip(lines[1])
	if !strings.HasPrefix(strings.TrimLeft(orders, " "), "×  orders") || !strings.Contains(orders, "● changed") {
		t.Fatalf("orders failed and changed: %q", orders)
	}
	if web := pane.Strip(lines[2]); strings.Contains(web, "changed") {
		t.Fatalf("web did not change: %q", web)
	}
	if kit := pane.Strip(lines[4]); !strings.Contains(kit, "–  kit") || !strings.Contains(kit, "go") {
		t.Fatalf("kit did not run and is written in go: %q", kit)
	}
}

func TestOnlyTheSelectedRowHasThePointer(t *testing.T) {
	lines, _, sel := ListBody(plain, sampleList(), 60)
	if sel != 2 {
		t.Fatalf("the selected line is %d, want 2 (web)", sel)
	}
	for i, l := range lines {
		if strings.Contains(l, Pointer) != (i == sel) {
			t.Fatalf("line %d: pointer %v, selected %v: %q", i, strings.Contains(l, Pointer), i == sel, l)
		}
	}
}

func TestTheCountLineSaysHowManyOfEachKindTheFilterKeeps(t *testing.T) {
	in := sampleList()
	if got := pane.Strip(ListHead(plain, in, 80)[1]); !strings.HasPrefix(got, "2 apps · 1 lib · 1 package · 1 changed") {
		t.Fatalf("the count line is %q", got)
	}
	in.Filter, in.Rows = "o", in.Rows[:1]
	head := ListHead(plain, in, 80)
	if got := pane.Strip(head[1]); !strings.HasPrefix(got, "1 of 2 apps · 0 of 1 lib · 0 of 1 package") {
		t.Fatalf("the filtered count line is %q", got)
	}
	if got := pane.Strip(head[0]); !strings.HasPrefix(got, "Filter: o") || strings.Contains(got, "▏") {
		t.Fatalf("the filter line is %q", got)
	}
	in.Typing = true
	if got := pane.Strip(ListHead(plain, in, 80)[0]); !strings.HasPrefix(got, "Filter: o▏") {
		t.Fatalf("while typing the filter line has the cursor: %q", got)
	}
}

func TestTheLegendSaysWhatEachMarkMeans(t *testing.T) {
	if got := pane.Strip(Legend(plain, 60)); !strings.HasPrefix(got, "✓ passed  × failed  – not run  ● changed") {
		t.Fatalf("legend %q", got)
	}
}

func TestNoLineOfTheListIsWiderThanItsWidth(t *testing.T) {
	in := sampleList()
	in.Rows[0].Name = "a-project-with-a-very-long-name-indeed"
	in.Rows[1].Version = "0.4.0 → 0.5.0"
	for w := 20; w <= 90; w++ {
		lines, _, _ := ListBody(plain, in, w)
		lines = append(lines, ListHead(plain, in, w)...)
		lines = append(lines, Legend(plain, w))
		for _, l := range lines {
			if got := pane.Width(l); got != w {
				t.Fatalf("w=%d: a line is %d cells: %q", w, got, pane.Strip(l))
			}
		}
	}
}

func TestAnEmptyFilterResultSaysHowToClearIt(t *testing.T) {
	in := sampleList()
	in.Rows, in.Filter = nil, "zzz"
	lines, rowOf, sel := ListBody(plain, in, 60)
	if len(lines) != 1 || len(rowOf) != 0 || sel != -1 || !strings.Contains(lines[0], "Esc clears it") {
		t.Fatalf("lines %q rows %v selected %d", lines, rowOf, sel)
	}
}

func TestPickedProjectsShowTheirTickBoxAndTheCountLeadsTheHead(t *testing.T) {
	in := sampleList()
	in.Rows[0].Picked, in.Rows[2].Picked = true, true
	in.Picked = 2
	lines, _, _ := ListBody(plain, in, 60)
	for i, want := range map[int]string{1: "[x] ×  orders", 2: "[ ] ✓  web", 4: "[x] –  kit", 6: "[ ] ✓  sdk"} {
		if got := pane.Strip(lines[i]); !strings.Contains(got, want) || pane.Width(lines[i]) != 60 {
			t.Fatalf("line %d is %q, want %q in 60 cells", i, got, want)
		}
	}
	if head := pane.Strip(ListHead(plain, in, 60)[1]); !strings.HasPrefix(head, "2 selected · 2 apps") {
		t.Fatalf("the count line leads with the selection: %q", head)
	}
	in.Picked = 0
	if lines, _, _ := ListBody(plain, in, 60); strings.Contains(pane.Strip(lines[1]), "[") {
		t.Fatalf("with nothing picked no row has a box: %q", pane.Strip(lines[1]))
	}
}
