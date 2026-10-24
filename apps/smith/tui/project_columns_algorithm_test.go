package tui

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
		c := columnsForProjects(tc.w, 20, 25, true)
		if c.kind != tc.kind || c.lang != tc.lang || c.stack != tc.stack {
			t.Errorf("w=%d: got %+v", tc.w, c)
		}
		if c.tasks < tasksFloorW && c.tasks < 25 {
			t.Errorf("w=%d: tasks squeezed to %d", tc.w, c.tasks)
		}
		used := projectLeadW + c.name + depsColW + c.tasks
		if c.kind {
			used += kindColW
		}
		if c.lang {
			used += langColW
		}
		if c.stack {
			used += stackColW
		}
		if used > tc.w {
			t.Errorf("w=%d: columns take %d", tc.w, used)
		}
	}
}

func TestAShortTasksListIsNotAReasonToDropColumns(t *testing.T) {
	if c := columnsForProjects(80, 10, 4, true); !c.stack || !c.lang || !c.kind {
		t.Errorf("got %+v", c)
	}
}

func TestTheStackColumnIsHiddenWhenNoProjectHasAStack(t *testing.T) {
	if c := columnsForProjects(200, 10, 20, false); c.stack || !c.lang || !c.kind {
		t.Errorf("got %+v", c)
	}
}
