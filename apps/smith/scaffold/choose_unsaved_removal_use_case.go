package scaffold

import (
	"bufio"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// The answers to "what happens to the files git cannot bring back".
const (
	removeKeepCopy = "keep"
	removeDelete   = "delete"
	removeCancel   = "cancel"
)

// chooseUnsaved lists the files that deleting the project folder would lose
// for good and asks what to do with them: move the folder to .smith/removed/
// (default), delete it, or cancel. Without a person at a terminal it refuses
// and names the flags, because smith does not pick between a backup and a
// deletion.
//
// Parameters: preview is the dry run of the removal; canAsk says a person can
// answer; in and out carry the question.
// Returns: removeKeepCopy or removeDelete; an error when cancelled or when
// nobody can answer.
func chooseUnsaved(preview Removal, canAsk bool, in *bufio.Reader, out io.Writer) (string, error) {
	if !canAsk {
		return "", problem.New(fmt.Sprintf("%s holds %d file(s) that git cannot bring back, and nobody can say what to do with them", preview.Project, len(preview.Unsaved)),
			"--keep-copy moves the folder to .smith/removed/ instead of deleting it", "--delete deletes it for good")
	}
	fmt.Fprintf(out, "\n%s holds %d file(s) that git cannot bring back:\n", preview.Root, len(preview.Unsaved))
	for i, f := range preview.Unsaved {
		if i == maxListed {
			fmt.Fprintf(out, "  and %d more\n", len(preview.Unsaved)-i)
			break
		}
		fmt.Fprintf(out, "  %s\n", f)
	}
	if preview.Ignored > 0 {
		fmt.Fprintf(out, "  (and %d ignored entries such as node_modules, which a build makes again)\n", preview.Ignored)
	}
	q := NewQuestion{
		Prompt:  "What should happen to the folder?",
		Default: removeKeepCopy, Flag: "keep-copy",
		Options: []NewOption{
			{Label: "Move it to .smith/removed/", Value: removeKeepCopy, Hint: "the project leaves the workspace and nothing is lost"},
			{Label: "Delete it for good", Value: removeDelete, Hint: "the files listed above cannot be brought back"},
			{Label: "Cancel", Value: removeCancel, Hint: "change nothing"},
		},
	}
	answer, _ := askNewQuestion(q, in, out)
	if answer == removeCancel {
		return "", operationplan.ErrDeclined
	}
	return answer, nil
}
