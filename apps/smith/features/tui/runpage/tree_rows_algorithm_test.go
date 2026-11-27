package runpage

import (
	"strings"
	"testing"
)

func treeText(rows []Row) string {
	var got []string
	for _, r := range rows {
		got = append(got, r.Project+"/"+r.TaskID)
	}
	return strings.Join(got, " ")
}

func TestTheTreeGroupsTheTasksByProjectInTheOrderTheyCame(t *testing.T) {
	rows := Tree([]TaskRef{{ID: "kit:build"}, {ID: "web:build"}, {ID: "kit:test", Failed: true}}, false)
	if got := treeText(rows); got != "kit/ kit/kit:build kit/kit:test web/ web/web:build" {
		t.Fatalf("tree %q", got)
	}
	if rows[1].Last || !rows[2].Last || !rows[4].Last {
		t.Fatalf("only the last task of a project closes its branch: %+v", rows)
	}
}

func TestErrorsOnlyKeepsTheFailedTasksAndTheirProjects(t *testing.T) {
	rows := Tree([]TaskRef{{ID: "kit:build"}, {ID: "web:build", Failed: true}, {ID: "kit:test"}}, true)
	if got := treeText(rows); got != "web/ web/web:build" {
		t.Fatalf("errors only %q", got)
	}
	if got := TaskIDs(rows); len(got) != 1 || got[0] != "web:build" {
		t.Fatalf("task ids %v", got)
	}
}

func TestATaskIDIsSplitIntoItsProjectAndItsTask(t *testing.T) {
	for id, want := range map[string][2]string{"web:test": {"web", "test"}, "verify": {"", "verify"}} {
		if ProjectOf(id) != want[0] || TaskOf(id) != want[1] {
			t.Errorf("%q: %q and %q", id, ProjectOf(id), TaskOf(id))
		}
	}
}
