// Package convertplan decides how an Nx workspace, or one made by the old
// smith, becomes a smith workspace: projects, kinds, tasks, release settings
// and the files to edit or remove. It reports everything it cannot map and
// changes nothing on disk.
package convertplan
