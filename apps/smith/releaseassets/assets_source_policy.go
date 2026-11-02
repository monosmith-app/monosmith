package releaseassets

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// The two sources a run for an existing tag can build from.
const (
	// SourceCurrent builds the checkout as it is, named with the tag's version.
	SourceCurrent = "current"
	// SourceTag builds the tag's own commit in a temporary git worktree.
	SourceTag = "tag"
	// SourceUndecided is the value the workflow input holds until someone picks
	// one. It means no source was given.
	SourceUndecided = "choose"
)

// chooseSource decides which code a run for an old tag builds. A source given
// with --source is checked and used. Without one, a terminal is asked, with the
// consequence of each option in plain words, and anything else is refused with
// both flags named: smith never picks the code on its own.
//
// It returns SourceCurrent or SourceTag. It returns a problem when the source
// is unknown, when no terminal can answer, or when the answer is not one of
// the two options.
func chooseSource(source, tag, version string, terminal bool, in io.Reader, out io.Writer) (string, error) {
	switch source {
	case SourceCurrent, SourceTag:
		return source, nil
	case "":
	default:
		return "", unknownSource(source)
	}
	if !terminal {
		return "", problem.New(fmt.Sprintf("%s is not at HEAD, and smith needs you to say which code to build", tag),
			"run it again with --source current (the code as it is now, named with the version "+version+") or --source tag (the tag's own commit, the exact code that was released)")
	}
	fmt.Fprintf(out, "%s is not at HEAD. Which code should its files be built from?\n", tag)
	fmt.Fprintf(out, "  1) current  the code as it is now, named with the version %s. The files carry newer code under an old number.\n", version)
	fmt.Fprintf(out, "  2) tag      the code of the tag's own commit, built in a temporary copy. This is the exact code that was released.\n")
	fmt.Fprint(out, "Choose 1 or 2: ")
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return "", problem.New("no answer was given for the source", "run it again with --source current or --source tag")
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "1", SourceCurrent:
		return SourceCurrent, nil
	case "2", SourceTag:
		return SourceTag, nil
	}
	return "", problem.New(fmt.Sprintf("%q is not one of the two options", strings.TrimSpace(line)), "answer 1 or 2, or run it again with --source current or --source tag")
}

// checkSource refuses a --source value that is neither empty, current nor tag.
func checkSource(source string) error {
	if source == "" || source == SourceCurrent || source == SourceTag {
		return nil
	}
	return unknownSource(source)
}

func unknownSource(source string) error {
	return problem.New(fmt.Sprintf("%q is not a source smith builds from", source), "use --source current or --source tag")
}
