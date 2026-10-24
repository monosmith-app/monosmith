package scaffold

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/commandargs"
	"github.com/spf13/cobra"
)

func TestAskMissingArgumentsAsksOnlyForWhatIsMissing(t *testing.T) {
	var out bytes.Buffer
	got, err := askMissingArguments([]string{"app"}, addQuestions(), strings.NewReader("Bad Name\nweb-site\n"), &out)
	if err != nil || strings.Join(got, " ") != "app web-site" {
		t.Fatalf("%v %v", got, err)
	}
	if strings.Contains(out.String(), "Kind") || !strings.Contains(out.String(), "must be kebab-case") {
		t.Errorf("only the name is asked, and a bad one is explained:\n%s", out.String())
	}
}

func TestAskMissingArgumentsTakesAKindByNumberOrCombinedName(t *testing.T) {
	got, err := askMissingArguments(nil, addQuestions(), strings.NewReader("2\napi\n"), &bytes.Buffer{})
	if err != nil || strings.Join(got, " ") != "lib api" {
		t.Fatalf("%v %v", got, err)
	}
	got, err = askMissingArguments(nil, addQuestions(), strings.NewReader("node-app\nweb\n"), &bytes.Buffer{})
	if err != nil || got[0] != "node-app" {
		t.Fatalf("%v %v", got, err)
	}
	got, err = askMissingArguments(nil, newQuestions(), strings.NewReader("acme\n"), &bytes.Buffer{})
	if err != nil || strings.Join(got, " ") != "acme" {
		t.Fatalf("%v %v", got, err)
	}
}

func TestAskMissingArgumentsGivesUpWithoutGuessing(t *testing.T) {
	if _, err := askMissingArguments(nil, newQuestions(), strings.NewReader("\n"), &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "needed") {
		t.Fatalf("an empty answer is not a name: %v", err)
	}
	if _, err := askMissingArguments([]string{"thing"}, addQuestions(), strings.NewReader(""), &bytes.Buffer{}); err == nil {
		t.Fatal("end of input is an error")
	}
	if got, err := askMissingArguments([]string{"app", "web"}, addQuestions(), strings.NewReader(""), &bytes.Buffer{}); err != nil || len(got) != 2 {
		t.Fatalf("nothing is asked when all is given: %v %v", got, err)
	}
}

func TestArgsOrAskKeepsTodaysMessageOutsideATerminal(t *testing.T) {
	spec := commandargs.Spec{Names: []string{"a kind", "a name"}, Example: "smith add app web"}
	cmd := &cobra.Command{Use: "add"}
	if err := argsOrAsk(spec, func() bool { return false })(cmd, []string{"app"}); err == nil || !strings.Contains(err.Error(), "needs a name") {
		t.Fatalf("outside a terminal the missing name is an error: %v", err)
	}
	if err := argsOrAsk(spec, func() bool { return true })(cmd, []string{"app"}); err != nil {
		t.Fatalf("in a terminal it is asked for later: %v", err)
	}
	if err := argsOrAsk(spec, func() bool { return true })(cmd, []string{"a", "b", "c"}); err == nil {
		t.Fatal("too many is still an error")
	}
}
