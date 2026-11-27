package commands

// countFor is how many projects a run of task on every project starts: the
// ones that have the task. Without a count per task it is every project.
func (s State) countFor(task string) int {
	if s.TaskProjects == nil {
		return s.Projects
	}
	return s.TaskProjects[task]
}

// allCount is the count an "all N" label shows for a run of task on every
// project, and whether it shows one: not before the projects are known, and
// not when no project has the task ("Lint all", never "Lint all 0").
func (s State) allCount(task string) (int, bool) {
	n := s.countFor(task)
	return n, s.Projects > 0 && n > 0
}
