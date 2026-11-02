package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// PrefixRequest is what ChooseModulePrefix needs.
type PrefixRequest struct {
	// Given is --module-prefix; when set it wins and nothing is asked.
	Given string
	// Name is the workspace name.
	Name string
	// Dir is the workspace folder, where git is asked for the remote and the user name.
	Dir string
	// Language is the main language, "" when undecided; only go or undecided asks.
	Language string
	// Interactive says a person can answer.
	Interactive bool
	// In and Out carry the question and the line that says what was chosen.
	In  *bufio.Reader
	Out io.Writer
}

var nonSlug = regexp.MustCompile(`[^a-z0-9-]+`)

// ChooseModulePrefix settles the Go module prefix every Go import path will
// carry. It is hard to change once code imports it, so smith does not write a
// made-up domain in silence. Order: --module-prefix; the origin remote; in a
// terminal, a question with three answers (github.com/<git user>/<name>,
// example.com/<name>, or one you type); otherwise example.com/<name>. In every
// case but the flag it prints what was chosen and how to change it. A project
// that is not Go (a language other than go or undecided) is not asked.
//
// Parameters: r is the request.
// Returns: the prefix; "" when the language is not Go and nothing was given.
func ChooseModulePrefix(r PrefixRequest) string {
	if r.Given != "" {
		return r.Given
	}
	git := gitclient.New(r.Dir)
	if git.IsRepository() {
		if p := remoteModulePrefix(git); p != "" {
			fmt.Fprintf(r.Out, "Go module prefix: %s (from the origin remote; --module-prefix changes it)\n", p)
			return p
		}
	}
	if r.Language != "" && r.Language != "go" {
		return ""
	}
	fallback := "example.com/" + r.Name
	if !r.Interactive {
		fmt.Fprintf(r.Out, "Go module prefix: %s (no git remote; --module-prefix <prefix> sets the real one before Go code imports it)\n", fallback)
		return fallback
	}
	options := []NewOption{}
	def := "example"
	if user := gitUserSlug(); user != "" {
		suggestion := "github.com/" + user + "/" + r.Name
		options = append(options, NewOption{Label: suggestion, Value: suggestion, Hint: "from your git user name; correct it if your account differs"})
		def = suggestion
	}
	options = append(options,
		NewOption{Label: fallback, Value: "example", Hint: "a placeholder: you rename it before you publish"},
		NewOption{Label: "Type one", Value: "type", Hint: "for example github.com/acme/" + r.Name},
	)
	q := NewQuestion{
		Prompt:  "This workspace has no git remote. Which Go module prefix should its projects use? Every Go import path will start with it.",
		Options: options, Default: def, Flag: "module-prefix",
	}
	answer, _ := askNewQuestion(q, r.In, r.Out)
	switch answer {
	case "example":
		answer = fallback
	case "type":
		free := NewQuestion{Prompt: "Type the module prefix", Default: fallback, Flag: "module-prefix", Check: nonEmpty("a module prefix")}
		answer, _ = askNewQuestion(free, r.In, r.Out)
	}
	fmt.Fprintf(r.Out, "Go module prefix: %s (change it with --module-prefix, or module_prefix in smith.toml before Go code imports it)\n", answer)
	return answer
}

// gitUserSlug is the git user name as a lowercase, URL-safe word, or "".
func gitUserSlug() string {
	name, err := gitclient.New(".").Run("config", "user.name")
	if err != nil {
		return ""
	}
	return strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-")), ""), "-")
}
