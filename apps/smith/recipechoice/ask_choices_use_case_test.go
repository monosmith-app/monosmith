package recipechoice

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

const teamRecipe = `name = "team"
version = "1.2.0"

[choices.language]
ask = "Language"
options = ["python", "go"]

[choices.kind]
default = "library"
options = ["library", "service"]

[when.language.python.use]
language = "python"

[when.language.go.use]
language = "go"

[when.language.go.choices.cgo]
ask = "Use cgo"
options = ["yes", "no"]
default = "no"

[when.language.go.settings.go]
cgo = "maybe"
`

func load(t *testing.T) recipe.Loaded {
	t.Helper()
	file := filepath.Join(t.TempDir(), "team.recipe.toml")
	if err := os.WriteFile(file, []byte(teamRecipe), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := (&recipe.Loader{}).Load(file)
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}

func asker(input string, interactive bool) (Asker, *bytes.Buffer) {
	var out bytes.Buffer
	return Asker{In: strings.NewReader(input), Out: &out, Interactive: interactive}, &out
}

func answer(res recipe.Resolution, name string) string {
	for _, q := range res.Asked {
		if q.Name == name {
			return q.Answer
		}
	}
	return "<not asked>"
}

func TestResolveWithEveryAnswerGivenAsksNothing(t *testing.T) {
	a, out := asker("", true)
	res, err := a.Resolve(load(t), map[string]string{"language": "python", "kind": "service"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 || answer(res, "language") != "python" || answer(res, "kind") != "service" {
		t.Fatalf("asked %q, answers %+v", out.String(), res.Asked)
	}
}

func TestResolveWithoutATerminalTakesDefaultsAndNamesWhatIsMissing(t *testing.T) {
	a, out := asker("", false)
	res, err := a.Resolve(load(t), map[string]string{"language": "go"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 || answer(res, "kind") != "library" || answer(res, "cgo") != "no" {
		t.Fatalf("asked %q, answers %+v", out.String(), res.Asked)
	}

	a, _ = asker("", false)
	_, err = a.Resolve(load(t), nil)
	var missing *recipe.MissingAnswerError
	if !errors.As(err, &missing) || !strings.Contains(err.Error(), `choice "language"`) || !strings.Contains(err.Error(), "--choice language=python|go") {
		t.Fatalf("want a missing answer naming language and the flag, got %v", err)
	}
}

func TestResolveAsksOneQuestionAtATimeDefaultsPreselected(t *testing.T) {
	// language: 2 (go); kind: Enter takes library; cgo (asked only for go): yes.
	a, out := asker("2\n\nyes\n", true)
	res, err := a.Resolve(load(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	if answer(res, "language") != "go" || answer(res, "kind") != "library" || answer(res, "cgo") != "yes" {
		t.Fatalf("answers %+v", res.Asked)
	}
	asked := out.String()
	for _, want := range []string{"Language (1 python, 2 go): ", "kind (1 library, 2 service) [library]: ", "Use cgo (1 yes, 2 no) [no]: "} {
		if !strings.Contains(asked, want) {
			t.Errorf("never asked %q in %q", want, asked)
		}
	}
	if strings.Index(asked, "Language") > strings.Index(asked, "Use cgo") {
		t.Errorf("the nested choice came before the one that adds it: %q", asked)
	}
}

func TestResolveDoesNotAskAChoiceItsBlockNeverAdded(t *testing.T) {
	a, out := asker("python\n\n", true)
	res, err := a.Resolve(load(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	if answer(res, "language") != "python" || answer(res, "cgo") != "<not asked>" || strings.Contains(out.String(), "cgo") {
		t.Fatalf("answers %+v, asked %q", res.Asked, out.String())
	}
}

func TestResolveAsksOnlyWhatWasNotGiven(t *testing.T) {
	a, out := asker("\n", true)
	if _, err := a.Resolve(load(t), map[string]string{"language": "python"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "Language") || !strings.Contains(out.String(), "kind") {
		t.Fatalf("asked %q", out.String())
	}
}

func TestResolveAsksAgainAfterAnInvalidAnswer(t *testing.T) {
	a, out := asker("perl\n7\n1\n\n", true)
	res, err := a.Resolve(load(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	if answer(res, "language") != "python" || strings.Count(out.String(), "choose one of: python, go") != 2 {
		t.Fatalf("answers %+v, asked %q", res.Asked, out.String())
	}
}

func TestResolveGivesUpAfterThreeInvalidAnswers(t *testing.T) {
	a, _ := asker("a\nb\nc\n", true)
	_, err := a.Resolve(load(t), nil)
	if err == nil || !strings.Contains(err.Error(), `"language"`) || !strings.Contains(err.Error(), "3 tries") {
		t.Fatalf("got %v", err)
	}
}

func TestResolveNamesTheChoiceWhenTheInputEnds(t *testing.T) {
	a, _ := asker("", true)
	_, err := a.Resolve(load(t), nil)
	if err == nil || !strings.Contains(err.Error(), `"language"`) || !strings.Contains(err.Error(), "--choice language=") {
		t.Fatalf("got %v", err)
	}
}

func TestResolveRejectsAnAnswerThatIsNotAnOption(t *testing.T) {
	a, _ := asker("", false)
	_, err := a.Resolve(load(t), map[string]string{"language": "perl"})
	if err == nil || !strings.Contains(err.Error(), "perl") {
		t.Fatalf("got %v", err)
	}
}
