package manifestoffer

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/shared/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
)

// Question is what Settle asks about a folder with sources and no manifest.
type Question struct {
	// Folder names the folder in the text ("the repository root", "../tools").
	Folder string
	// Finds are the languages whose sources are there (Sources).
	Finds []Find
	// Answer is --manifest: a language of Finds, None, or "" when not given.
	Answer string
	// Ask is true when a person can answer in the terminal.
	Ask bool
	// Command is the command the remedy repeats with --manifest
	// ("smith import ../tools").
	Command string
	// In and Out are the terminal.
	In  *bufio.Reader
	Out io.Writer
}

// Settle decides whether smith writes a manifest, and for which language:
// --manifest answers; else a terminal asks, with the language that has the
// most files as the default; else smith refuses, because it never writes a
// manifest nobody asked for. Nothing is changed.
//
// Parameters: q is the question.
// Returns: the chosen find and true, or false when the answer is none, or a
// problem for an answer that is not a choice and for no terminal and no flag.
func Settle(q Question) (Find, bool, error) {
	if len(q.Finds) == 0 {
		return Find{}, false, nil
	}
	if q.Answer != "" {
		if strings.EqualFold(q.Answer, None) {
			return Find{}, false, nil
		}
		for _, f := range q.Finds {
			if strings.EqualFold(f.Language, q.Answer) {
				return f, true, nil
			}
		}
		return Find{}, false, problem.New(fmt.Sprintf("--manifest %s is not a choice: %s has sources of %s", q.Answer, q.Folder, languages(q.Finds)),
			fmt.Sprintf("use --manifest %s, or --manifest %s to write none", q.Finds[0].Language, None))
	}
	first := q.Finds[0]
	if !q.Ask {
		return Find{}, false, problem.New(fmt.Sprintf("%s has %s and no manifest, and smith writes one only when you ask", q.Folder, describe(first)),
			fmt.Sprintf("%s --manifest %s writes %s from the default stack of %s; the sources stay where they are", q.Command, first.Language, first.Manifest, first.Language),
			"or write the manifest yourself, then run the command again")
	}
	options := make([]choiceprompt.Option, 0, len(q.Finds)+1)
	for _, f := range q.Finds {
		options = append(options, choiceprompt.Option{Label: "write " + f.Manifest + " (" + f.Language + ")", Value: f.Language,
			Hint: describe(f) + "; the default stack's manifest, the sources stay where they are"})
	}
	options = append(options, choiceprompt.Option{Label: "none", Value: None, Hint: "stop, and write the manifest yourself"})
	prompt := fmt.Sprintf("%s has source files and no manifest, so smith sees no project there. Write one?", capital(q.Folder))
	answer := choiceprompt.Ask(prompt, options, first.Language, q.In, q.Out)
	for _, f := range q.Finds {
		if f.Language == answer {
			return f, true, nil
		}
	}
	return Find{}, false, nil
}

// describe says how many source files of the language were found, with the first few.
func describe(f Find) string {
	shown := f.Files
	if len(shown) > 3 {
		shown = shown[:3]
	}
	more := ""
	if len(f.Files) > len(shown) {
		more = ", ..."
	}
	return fmt.Sprintf("%d %s source file(s) (%s%s)", len(f.Files), f.Language, strings.Join(shown, ", "), more)
}

func languages(finds []Find) string {
	names := make([]string, len(finds))
	for i, f := range finds {
		names[i] = f.Language
	}
	return strings.Join(names, ", ")
}

func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
