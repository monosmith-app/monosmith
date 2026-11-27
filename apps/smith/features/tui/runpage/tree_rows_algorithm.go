package runpage

import "strings"

// TaskRef is what the tree needs of a task: its ID ("web:test") and whether it
// failed.
type TaskRef struct {
	ID     string
	Failed bool
}

// Row is one row of the tree: the heading of a project (TaskID is empty) or a
// task under it. Last is set on the last task of its project, which is drawn
// with the closing branch.
type Row struct {
	Project string
	TaskID  string
	Last    bool
}

// Heading is true for the row that names a project.
func (r Row) Heading() bool { return r.TaskID == "" }

// ProjectOf is the project part of a task ID ("web" of "web:test"); an ID
// without a colon has no project.
//
// Parameters: id is the task ID.
// Returns: the project, or "" when the ID names none.
func ProjectOf(id string) string {
	project, _, ok := strings.Cut(id, ":")
	if !ok {
		return ""
	}
	return project
}

// TaskOf is the task part of a task ID ("test" of "web:test"), or the whole ID
// when it names no project.
//
// Parameters: id is the task ID.
// Returns: the task name.
func TaskOf(id string) string {
	if _, task, ok := strings.Cut(id, ":"); ok {
		return task
	}
	return id
}

// Tree groups the tasks by project, the projects in the order their first
// task comes. Each project is a heading row followed by its tasks. With
// errorsOnly set only the failed tasks are kept, and a project without one is
// left out. A task without a project sits under a heading with no name.
//
// Parameters: tasks are the tasks of the run in their order; errorsOnly keeps
// only the failed ones.
// Returns: the rows, headings included.
func Tree(tasks []TaskRef, errorsOnly bool) []Row {
	var order []string
	groups := map[string][]string{}
	for _, t := range tasks {
		if errorsOnly && !t.Failed {
			continue
		}
		p := ProjectOf(t.ID)
		if _, seen := groups[p]; !seen {
			order = append(order, p)
		}
		groups[p] = append(groups[p], t.ID)
	}
	var out []Row
	for _, p := range order {
		out = append(out, Row{Project: p})
		ids := groups[p]
		for i, id := range ids {
			out = append(out, Row{Project: p, TaskID: id, Last: i == len(ids)-1})
		}
	}
	return out
}

// TaskIDs are the task IDs of the rows, in their order: the stops of the keys
// that move the selection.
//
// Parameters: rows are the rows of a tree.
// Returns: the IDs of the task rows.
func TaskIDs(rows []Row) []string {
	var ids []string
	for _, r := range rows {
		if !r.Heading() {
			ids = append(ids, r.TaskID)
		}
	}
	return ids
}
