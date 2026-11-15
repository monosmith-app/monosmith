package release

// Level is how much a version moves.
type Level int

// Levels, ordered.
const (
	None Level = iota
	Patch
	Minor
	Major
)

func (l Level) String() string {
	return [...]string{"none", "patch", "minor", "major"}[l]
}

// LevelOf decides what one commit asks for: a breaking change is major, feat
// is minor, fix and perf are patch, and every other type (docs, chore, test,
// ci, refactor, build, style) releases nothing.
func LevelOf(c Commit) Level {
	if !c.Valid || IsReleaseCommit(c) {
		return None
	}
	if c.Breaking {
		return Major
	}
	switch c.Type {
	case "feat":
		return Minor
	case "fix", "perf":
		return Patch
	}
	return None
}

// LevelOfAll is the highest level any commit asks for.
func LevelOfAll(commits []Commit) Level {
	l := None
	for _, c := range commits {
		l = max(l, LevelOf(c))
	}
	return l
}
