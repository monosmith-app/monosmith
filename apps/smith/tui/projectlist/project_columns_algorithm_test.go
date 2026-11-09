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
		c := ColumnsFor(tc.w, 20, true)
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
	if c := ColumnsFor(200, 12, true); c.Name != 12 {
		t.Errorf("a wide table gives the name %d, want 12", c.Name)
	}
}

func TestTheStackColumnIsHiddenWhenNoProjectHasAStack(t *testing.T) {
	if c := ColumnsFor(200, 10, false); c.Stack || !c.Lang || !c.Kind {
		t.Errorf("got %+v", c)
	}
}
