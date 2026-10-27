package cipipeline

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/affected"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// nativeExcluded leaves the native projects to the native job, on every OS.
var nativeExcluded = []string{"!tag:" + release.NativeTag}

// verifyIntro says what verify is about to check: how many projects, which
// tasks, and for a pull request why only some.
func verifyIntro(s *session.Session, base string) []string {
	names, _ := s.Graph.Select(nativeExcluded)
	total := len(names)
	line := fmt.Sprintf("Checking %s (%s).", plural(total, "project"), strings.Join(VerifyTasks, ", "))
	if base != "" {
		res := affected.Compute(s.Git, s.Graph, base)
		keep := map[string]bool{}
		for _, n := range res.Projects {
			keep[n] = true
		}
		n := 0
		for _, name := range names {
			if keep[name] {
				n++
			}
		}
		line = fmt.Sprintf("Checking the %d of %s that this change affects (%s), with %s.", n, plural(total, "project"), res.Reason, strings.Join(VerifyTasks, ", "))
	}
	lines := []string{line, "The dependency boundaries between projects are checked first."}
	return lines
}

// packIntro says what pack is about to build and where it goes.
func packIntro(s *session.Session) []string {
	names, _ := s.Graph.Select(nativeExcluded)
	packing := 0
	for _, n := range names {
		if _, ok := s.Graph.Get(n).Tasks[PackTask]; ok {
			packing++
		}
	}
	return []string{fmt.Sprintf("Packing %s (task %q) into %s.", plural(packing, "project"), PackTask, DropDir)}
}
