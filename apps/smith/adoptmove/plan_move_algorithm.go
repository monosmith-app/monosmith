package adoptmove

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// PlanMove works out the git mv commands that put the project in dir. files are
// the repository's files (git ls-files), repository-relative with forward
// slashes; dir is a folder CleanFolder accepted. Every top-level entry moves
// whole, except the ones Classify keeps at the root. It changes nothing.
//
// It returns a problem, and no plan, when the layout is ambiguous: the folder
// already holds files, the folder's first part is a file of the repository or a
// folder of the project that moves, or the manifest is not in the files.
func PlanMove(c Candidate, dir string, files []string) (Plan, error) {
	top := TopLevel(files)
	hasManifest := false
	for _, name := range top {
		hasManifest = hasManifest || name == c.Manifest
	}
	if !hasManifest {
		return Plan{}, problem.New(c.Manifest+" is not in the repository's files, so smith does not know what to move",
			"commit "+c.Manifest+" (git add, git commit), then run smith adopt again")
	}
	first, _, _ := strings.Cut(dir, "/")
	for _, f := range files {
		switch {
		case f == dir || strings.HasPrefix(f, dir+"/"):
			return Plan{}, problem.New(dir+" already holds files, and moving the project in would mix the two", "choose a folder that does not exist yet")
		case f == first:
			return Plan{}, problem.New(fmt.Sprintf("%s is a file, so %s cannot be a folder", first, dir), "choose another folder")
		}
	}
	p := Plan{Dir: dir, Ecosystem: c.Ecosystem}
	for _, name := range top {
		if Classify(c.Ecosystem, name) == RoleRepository {
			p.Stays = append(p.Stays, name)
			continue
		}
		if name == first {
			return Plan{}, problem.New(fmt.Sprintf("the project has a folder or file named %s, and %s is inside that name", first, dir),
				"choose a folder whose first part is not "+first+", or rename it first (git mv)")
		}
		p.Moves = append(p.Moves, Step{From: name, To: dir + "/" + name})
	}
	sort.Slice(p.Moves, func(i, j int) bool { return p.Moves[i].From < p.Moves[j].From })
	p.Notes = notesFor(c, top)
	return p, nil
}

func notesFor(c Candidate, top []string) []string {
	if c.Ecosystem != EcosystemNode {
		return nil
	}
	var notes []string
	has := map[string]bool{}
	for _, name := range top {
		has[name] = true
	}
	if has["package-lock.json"] {
		notes = append(notes, "package-lock.json stays at the root: with npm workspaces the one lockfile lives there, and npm install rewrites it for the new layout.")
	}
	for _, other := range []string{"yarn.lock", "pnpm-lock.yaml", "bun.lockb", "bun.lock"} {
		if has[other] {
			notes = append(notes, other+" stays at the root: smith installs with npm, so run npm install to write package-lock.json (smith convert does this for a yarn or pnpm repository).")
		}
	}
	return notes
}

// Lines are the plan as the commands a person would type, one per line, for a
// dry run.
func (p Plan) Lines() []string {
	out := make([]string, 0, len(p.Moves))
	for _, m := range p.Moves {
		out = append(out, "git mv "+m.From+" "+m.To)
	}
	return out
}
