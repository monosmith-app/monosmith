package recipewizard

import (
	"strings"
	"testing"
)

const answersFile = `
name = "team"
version = "0.1.0"

[choices.language]
ask = "Language"
options = ["python", "node"]
default = "python"

[levels.language]
choice = "language"
plugins = { python = "python", node = "node" }

[levels.lint]
choice = "language"
plugins = { python = "ruff", node = "eslint" }

[levels.structure]
fixed = ["slices"]
`

func TestReadPlanReadsAnAnswersFile(t *testing.T) {
	p, err := ReadPlan([]byte(answersFile), "answers.toml")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "team" || p.Levels["lint"].PluginsOf("node")[0] != "eslint" || p.Levels["structure"].Fixed[0] != "slices" || p.Choices["language"].Default != "python" {
		t.Errorf("plan = %+v", p)
	}
}

func TestReadPlanRefusesWhatARecipeCouldNotUse(t *testing.T) {
	for name, tc := range map[string]struct{ edit, want string }{
		"unknown key":           {"\nflavour = 1\n", "flavour"},
		"a name a tag can't be": {"name = \"Team\"", "lowercase words"},
		"a version":             {"version = \"1\"", "major.minor.patch"},
		"both ways":             {"[levels.test]\nfixed = [\"pytest\"]\nchoice = \"language\"\nplugins = { python = \"pytest\" }", "alternatives"},
		"neither":               {"[levels.test]\n", "give fixed plugins"},
		"two languages":         {"[levels.scaffold]\nfixed = [\"node\", \"python\"]", "takes one plugin"},
		"an answer not offered": {"[levels.test]\nchoice = \"language\"\nplugins = { ruby = \"rspec\" }", "not an option"},
		"one option":            {"[choices.kind]\noptions = [\"app\"]", "at least two options"},
	} {
		text := answersFile + "\n" + tc.edit
		if strings.HasPrefix(tc.edit, "name") || strings.HasPrefix(tc.edit, "version") {
			text = strings.Replace(answersFile, strings.SplitN(tc.edit, " =", 2)[0]+" = "+firstValue(answersFile, strings.SplitN(tc.edit, " =", 2)[0]), tc.edit, 1)
		}
		_, err := ReadPlan([]byte(text), "answers.toml")
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: err = %v, want it to mention %q", name, err, tc.want)
		}
	}
}

func firstValue(text, key string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, key+" = ") {
			return strings.TrimPrefix(line, key+" = ")
		}
	}
	return ""
}
