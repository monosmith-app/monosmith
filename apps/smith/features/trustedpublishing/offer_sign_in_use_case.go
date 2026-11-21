package trustedpublishing

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/apps/smith/shared/choiceprompt"
)

// Answer values of the question.
const (
	answerTrusted = "trusted"
	answerKey     = "key"
)

// Question is what Offer needs.
type Question struct {
	// Root is the workspace root, where smith.toml is.
	Root string
	// Config is the loaded configuration; Offer records each answer in its
	// [release] section too, so the caller generates with it.
	Config *workspace.Config
	// Secrets are the credential names the generated pipelines pass.
	Secrets []string
	// Place is where the workflow runs, for the set-up steps.
	Place Place
	// Ask is true when a person can answer in the terminal (not --yes, not CI).
	Ask bool
	// In and Out are the terminal; Out also takes the notes of a run that
	// does not ask.
	In  *bufio.Reader
	Out io.Writer
}

// Offer settles each pending keyless sign-in (Pending). With a person at the
// terminal it explains the one-time set-up and asks: trusted publishing (the
// default, as it keeps no long-lived key) or the API key secret, then writes
// the answer to smith.toml, false for the key, so the question is not asked
// again. Without a person it changes nothing and prints one note per option
// that names the setting.
//
// Parameters: q is the question.
// Returns: the options it recorded, or an error when smith.toml cannot be
// written.
func Offer(q Question) ([]Option, error) {
	var recorded []Option
	for _, o := range Pending(*q.Config, q.Secrets) {
		if !q.Ask {
			fmt.Fprintf(q.Out, "note: %s publishes with the secret %s.\n→ smith config set %s true signs in with %s instead, false keeps the key and stops this note\n", capital(o.Registry), o.Secret, o.Path(), o.Method)
			continue
		}
		trusted := ask(o, q) == answerTrusted
		if err := configedit.Set(q.Root, o.Path(), []string{strconv.FormatBool(trusted)}); err != nil {
			return recorded, err
		}
		o.record(&q.Config.Release, trusted)
		recorded = append(recorded, o)
		fmt.Fprintf(q.Out, "wrote %s: [release] %s = %t\n", workspace.FileName, o.Setting, trusted)
	}
	return recorded, nil
}

// ask explains the option and returns the answer.
func ask(o Option, q Question) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: the pipeline publishes with the secret %s, a long-lived key.\n", capital(o.Registry), o.Secret)
	fmt.Fprintf(&b, "With %s the workflow signs in with its own short-lived token instead. Set up once:\n", o.Method)
	for _, step := range o.Steps(q.Place) {
		fmt.Fprintf(&b, "  - %s\n", step)
	}
	fmt.Fprintf(&b, "How does the pipeline sign in to %s?", o.Registry)
	options := []choiceprompt.Option{
		{Label: o.Method, Value: answerTrusted, Hint: "recommended: no long-lived key to store or renew (" + o.Setting + " = true)"},
		{Label: "API key secret", Value: answerKey, Hint: "keep the repository secret " + o.Secret + " (" + o.Setting + " = false)"},
	}
	return choiceprompt.Ask(b.String(), options, answerTrusted, q.In, q.Out)
}

func capital(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
