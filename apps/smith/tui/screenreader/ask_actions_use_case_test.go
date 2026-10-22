package screenreader

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
)

func sampleActions() []guide.Action {
	return guide.Actions([]string{"go", "node"}, []string{"go", "node"}, []string{"go"})
}

func converse(t *testing.T, input string) (argv []string, output string, err error) {
	t.Helper()
	var out strings.Builder
	argv, err = Ask(strings.NewReader(input), &out, sampleActions())
	return argv, out.String(), err
}

func TestAnActionWithQuestionsEndsAsItsCommandLine(t *testing.T) {
	// 3 is "Add a project": kind (2: lib), language (2: node), name.
	argv, out, err := converse(t, "3\n2\n2\ncore\n")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"add", "lib", "core", "--lang", "node"}; !reflect.DeepEqual(argv, want) {
		t.Fatalf("argv %v, want %v\n%s", argv, want, out)
	}
	for _, want := range []string{"2. Run tasks - build, test, lint, typecheck", "3. Add a project - app, lib or package", "What kind? Type the number of one.", "2. lib - internal, never published", "Which language? Type the number of one.", "Name (kebab-case). For example web."} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestAMultipleChoiceTakesSeveralNumbers(t *testing.T) {
	argv, _, err := converse(t, "2\n1, 3\n2\n")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"run", "lint", "test", "--affected"}; !reflect.DeepEqual(argv, want) {
		t.Fatalf("argv %v, want %v", argv, want)
	}
}

func TestAnActionWithoutQuestionsIsRunAtOnce(t *testing.T) {
	argv, _, err := converse(t, "5\n")
	if err != nil || !reflect.DeepEqual(argv, []string{"graph"}) {
		t.Fatalf("argv %v err %v", argv, err)
	}
}

func TestQuittingAnywhereRunsNothing(t *testing.T) {
	for _, input := range []string{"q\n", "", "3\nq\n", "3\n1\n2\n", "2\n1\nQ\n"} {
		argv, _, err := converse(t, input)
		if err != nil || argv != nil {
			t.Fatalf("%q: argv %v err %v", input, argv, err)
		}
	}
}

func TestAWrongNumberIsSaidAndAskedAgainThenGivenUp(t *testing.T) {
	argv, out, err := converse(t, "nine\n0\n42\n")
	if err == nil || argv != nil {
		t.Fatalf("argv %v err %v", argv, err)
	}
	if want := fmt.Sprintf(`"nine" is not one of the numbers 1 to %d.`, len(sampleActions())); !strings.Contains(out, want) {
		t.Fatalf("the mistake was not said:\n%s", out)
	}
	argv, out, err = converse(t, "3\n9\n1\n1\nweb\n")
	if err != nil || !reflect.DeepEqual(argv, []string{"add", "app", "web", "--lang", "go"}) {
		t.Fatalf("a mistake is asked again: %v %v\n%s", argv, err, out)
	}
	if !strings.Contains(out, `"9" is not one of the numbers 1 to 3.`) {
		t.Fatalf("%s", out)
	}
}

func TestTheOutputIsPlainText(t *testing.T) {
	_, out, _ := converse(t, "3\n1\n1\nweb\n")
	for _, r := range out {
		if r == '\x1b' || r > 0x7e {
			t.Fatalf("a screen reader would read %q aloud:\n%s", r, out)
		}
	}
}

func TestAnOptionalQuestionCanBeSkippedWithEnter(t *testing.T) {
	action := guide.Action{
		Title: "Add a pipeline",
		Steps: []guide.Step{{Key: "template", Prompt: "template (optional)", Optional: true}},
		Argv: func(a map[string][]string) []string {
			if len(a["template"]) == 0 || a["template"][0] == "" {
				return []string{"pipeline"}
			}
			return []string{"pipeline", a["template"][0]}
		},
	}
	var out strings.Builder
	argv, err := Ask(strings.NewReader("1\n\n"), &out, []guide.Action{action})
	if err != nil || len(argv) != 1 || argv[0] != "pipeline" {
		t.Fatalf("argv %v err %v\n%s", argv, err, out.String())
	}
	if !strings.Contains(out.String(), "Press Enter to skip.") {
		t.Errorf("the question does not say it can be skipped:\n%s", out.String())
	}
}
