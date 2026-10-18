package recipe

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// check runs `smith recipe check` over a recipe in a directory that is no
// workspace, and returns what it printed.
func check(t *testing.T, text string, asJSON bool, pairs ...string) (stdout, stderr string, err error) {
	t.Helper()
	dir := t.TempDir()
	path := writeFile(t, dir, "team.toml", strings.TrimPrefix(text, "\n"))
	t.Chdir(dir)
	answers, perr := parseAnswers(pairs)
	if perr != nil {
		t.Fatal(perr)
	}
	var out, errOut bytes.Buffer
	err = runCheck(&out, &errOut, path, answers, asJSON, "1.0.0")
	return out.String(), errOut.String(), err
}

func TestCheckPrintsWhatTheRecipeWouldUseAndWhereFrom(t *testing.T) {
	out, errOut, err := check(t, teamRecipe, false, "language=csharp", "kind=service")
	if err != nil || errOut != "" {
		t.Fatalf("err = %v, stderr = %q", err, errOut)
	}
	for _, want := range []string{
		"Recipe team 1.2.0",
		"language     csharp         (answered; options python, csharp) Language?",
		"kind         service        (answered; options library, service)",
		"language     dotnet",
		":20 (recipe team)",
		"deploy       docker",
		"structure    vertical-slices",
		"dotnet.target_framework",
		"net10.0",
		"The recipe is valid.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output must contain %q:\n%s", want, out)
		}
	}
}

func TestCheckTellsWhichChoicesNeedAnAnswer(t *testing.T) {
	out, errOut, err := check(t, teamRecipe, false)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(out, "needs an answer") || strings.Contains(out, "The recipe is valid.") {
		t.Errorf("stdout:\n%s", out)
	}
	if !strings.Contains(errOut, "error: choice \"language\"") || !strings.Contains(errOut, "--choice language=python|csharp") {
		t.Errorf("stderr:\n%s", errOut)
	}
}

func TestCheckReportsAPluginThatIsNotThere(t *testing.T) {
	_, errOut, err := check(t, "\nname = \"r\"\n[use]\nlanguage = \"python\"\nlint = \"ruff\"\n", false)
	if !errors.Is(err, ErrInvalid) || !strings.Contains(errOut, `use.lint names the plugin "ruff"`) || !strings.Contains(errOut, ":4 (recipe r)") {
		t.Errorf("err = %v, stderr:\n%s", err, errOut)
	}
}

func TestCheckPrintsJSONForTools(t *testing.T) {
	out, _, err := check(t, teamRecipe, true, "language=python")
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Name    string
		Chain   []Link
		Choices []Question
		Use     []Use
		Errors  []string
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("%v:\n%s", err, out)
	}
	if got.Name != "team" || len(got.Chain) != 1 || len(got.Choices) != 2 || len(got.Use) != 2 || len(got.Errors) != 0 {
		t.Errorf("report = %+v", got)
	}

	out, _, err = check(t, teamRecipe, true)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err = %v", err)
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil || len(got.Errors) == 0 {
		t.Errorf("a recipe that is not valid must say why in the JSON: %v\n%s", err, out)
	}
}

func TestParseAnswersWantsNameAndAnswer(t *testing.T) {
	if got, err := parseAnswers([]string{"a=b", "c=d=e"}); err != nil || got["a"] != "b" || got["c"] != "d=e" {
		t.Errorf("answers = %v, %v", got, err)
	}
	for _, bad := range []string{"a", "=b", "a="} {
		if _, err := parseAnswers([]string{bad}); err == nil {
			t.Errorf("%q must be refused", bad)
		}
	}
}
