package fileedits

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RetargetDependents rewrites, in the other projects of one language, the
// references to the directory of the project that changes kind, so they name
// its new directory.
//
// language is the plugin's language; a project of another language is left
// alone, and so is the changing project itself. files returns the
// workspace-relative files of a dependent project's root that may hold
// references; a file that does not exist is skipped. Every reference is
// rewritten with pathrefs.Retarget, from the directory of its project.
func RetargetDependents(edits *Set, params protocol.ChangeKindParams, language string, files func(root string) []string) {
	for _, other := range params.Projects {
		if other.Language != language || other.Root == params.Project.Root {
			continue
		}
		for _, file := range files(other.Root) {
			if text, ok := edits.Read(file); ok {
				updated, _ := pathrefs.Retarget(text, other.Root, params.Project.Root, params.NewRoot)
				edits.Write(file, updated)
			}
		}
	}
}
