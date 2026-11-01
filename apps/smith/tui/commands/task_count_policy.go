package commands

// countFor is how many projects a run of task on every project starts: the
// ones that have the task. Without a count per task it is every project.
func (s State) countFor(task string) int {
	if s.TaskProjects == nil {
		return s.Projects
	}
	return s.TaskProjects[task]
}
