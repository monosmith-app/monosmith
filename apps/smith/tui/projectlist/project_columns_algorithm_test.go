package projectlist

import "testing"

func TestTheProjectTableDropsOptionalColumnsWhenTheWidthIsShort(t *testing.T) {
	for _, tc := range []struct {
		w                 int
		kind, lang, stack bool
	}{
		{140, true, true, true},
		{64, true, true, true},
		{40, true, false, false},
		{24, false, false, false},
	} {
		c := ColumnsFor(tc.w, 20, true, false)
		if c.Kind != tc.kind || c.Lang != tc.lang || c.Stack != tc.stack {
			t.Errorf("w=%d: got %+v", tc.w, c)
		}
		used := LeadW + c.Name + DepsW
		if c.Kind {
			used += KindW
		}
		if c.Lang {
			used += LangW
		}
		if c.Stack {
			used += StackW
		}
		if used > tc.w {
			t.Errorf("w=%d: columns take %d", tc.w, used)
		}
	}
}

func TestTheNameColumnTakesTheRoomItWantsAndNoMore(t *testing.T) {
	if c := ColumnsFor(200, 12, true, false); c.Name != 12 {
		t.Errorf("a wide table gives the name %d, want 12", c.Name)
	}
}

func TestTheStackColumnIsHiddenWhenNoProjectHasAStack(t *testing.T) {
	if c := ColumnsFor(200, 10, false, false); c.Stack || !c.Lang || !c.Kind {
		t.Errorf("got %+v", c)
	}
}

func TestTheVersionColumnIsTheLastOptionalOneToGo(t *testing.T) {
	if c := ColumnsFor(140, 20, true, true); !c.Version || !c.Kind || !c.Lang || !c.Stack {
		t.Errorf("a wide table keeps every column: %+v", c)
	}
	if c := ColumnsFor(60, 20, false, true); !c.Version || c.Stack {
		t.Errorf("a short table drops the others before the version: %+v", c)
	}
	if c := ColumnsFor(140, 20, true, false); c.Version {
		t.Errorf("no project has a version, so no column: %+v", c)
	}
}

func TestAPendingReleaseShowsAsCurrentArrowNextAndShortensToTheArrowAlone(t *testing.T) {
	for _, tc := range []struct {
		current, next string
		w             int
		want          string
	}{
		{"0.4.0", "", 14, "0.4.0"},
		{"0.4.0", "0.4.0", 14, "0.4.0"},
		{"0.4.0", "0.5.0", 14, "0.4.0 → 0.5.0"},
		{"0.4.0", "0.5.0", 8, "→ 0.5.0"},
		{"", "1.0.0", 14, "→ 1.0.0"},
	} {
		if got := VersionText(tc.current, tc.next, tc.w); got != tc.want {
			t.Errorf("VersionText(%q, %q, %d) = %q, want %q", tc.current, tc.next, tc.w, got, tc.want)
		}
	}
}
