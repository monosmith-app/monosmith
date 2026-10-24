package projectlist

import "testing"

func TestTheProjectTableDropsOptionalColumnsBeforeCuttingTasks(t *testing.T) {
	for _, tc := range []struct {
		w                 int
		kind, lang, stack bool
	}{
		{140, true, true, true},
		{64, true, false, false},
		{50, false, false, false},
	} {
		c := ColumnsFor(tc.w, 20, 25, true)
		if c.Kind != tc.kind || c.Lang != tc.lang || c.Stack != tc.stack {
			t.Errorf("w=%d: got %+v", tc.w, c)
		}
		if c.Tasks < TasksFloorW && c.Tasks < 25 {
			t.Errorf("w=%d: tasks squeezed to %d", tc.w, c.Tasks)
		}
		used := LeadW + c.Name + DepsW + c.Tasks
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

func TestAShortTasksListIsNotAReasonToDropColumns(t *testing.T) {
	if c := ColumnsFor(80, 10, 4, true); !c.Stack || !c.Lang || !c.Kind {
		t.Errorf("got %+v", c)
	}
}

func TestTheStackColumnIsHiddenWhenNoProjectHasAStack(t *testing.T) {
	if c := ColumnsFor(200, 10, 20, false); c.Stack || !c.Lang || !c.Kind {
		t.Errorf("got %+v", c)
	}
}
