package scaffold

import (
	"fmt"
	"strconv"

	"github.com/monosmith-app/monosmith/apps/smith/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"
)

// The answers smith new keeps when nothing else is said.
const (
	DefaultReleaseBranch = "main"
	DefaultReleaseRemote = "origin"
	DefaultCacheDir      = ".smith/cache"
	DefaultCacheMaxMB    = 2048
	// DefaultNodeStack is the stack Node projects get unless another is chosen.
	DefaultNodeStack = "node-jest"
	// DefaultArchitecture is the code layout smith new turns on.
	DefaultArchitecture = "vertical-slices"
	// LanguageLater is the answer of "decide later" to the main language.
	LanguageLater = "later"
)

// Keys of the questions smith new asks. The CLI and the TUI form take their
// questions from NewQuestions, so both ask the same ones; each has a flag of
// smith new.
const (
	AnswerCI             = "ci"
	AnswerLanguage       = "language"
	AnswerNodeStack      = "node-stack"
	AnswerArchitecture   = "architecture"
	AnswerMore           = "more"
	AnswerReleaseBranch  = "release-branch"
	AnswerReleaseRemote  = "release-remote"
	AnswerGitHubReleases = "github-releases"
	AnswerCacheDir       = "cache-dir"
	AnswerCacheMaxMB     = "cache-max-mb"
	AnswerSaveRecipe     = "save-recipe"
	AnswerRecipeName     = "recipe-name"
)

// MoreYes is the answer of "more options" that asks the advanced questions.
const MoreYes = "yes"

// NewOption is one answer to a question.
type NewOption struct{ Label, Value, Hint string }

// NewQuestion is one question of smith new. With no Options it is free text
// and an empty answer takes Default.
type NewQuestion struct {
	Key    string
	Prompt string
	// Options are the answers on offer, the default first; nil for free text.
	Options []NewOption
	// Default is the Value (or text) Enter takes.
	Default string
	// Flag is the smith new flag that answers it without asking.
	Flag string
	// AtEnd questions are asked last. The command asks them after the summary
	// (askSaveRecipe), the form as its final steps.
	AtEnd bool
	// Advanced questions are asked only after "more options".
	Advanced bool
	// Applies says whether the question is asked, given the answers so far;
	// nil means always.
	Applies func(answers map[string]string) bool
	// Check rejects a free-text answer.
	Check func(value string) error
}

// NodeStacks are the stacks smith new offers Node projects, the default first.
func NodeStacks() []NewOption {
	return []NewOption{
		{Label: "Jest", Value: "node-jest", Hint: "TypeScript, ESLint and Jest (default)"},
		{Label: "Vitest", Value: "node-vitest", Hint: "TypeScript, ESLint and Vitest"},
		{Label: "TypeScript only", Value: "node-ts", Hint: "tsc, no lint and no tests"},
	}
}

// layoutHints say each code layout in a few words, short enough for one line
// beside its name (smith architecture says more).
var layoutHints = map[string]string{
	"vertical-slices":  "a folder per outcome, role-named files",
	"hexagonal":        "ports and adapters",
	"clean":            "dependencies point inward",
	"ddd":              "a folder per bounded context",
	"modular-monolith": "modules used only through their public API",
	"mvc":              "models, views and controllers",
	"package-by-layer": "controller, service and repository layers",
	"none":             "no layout checks",
}

// NewQuestions are the questions of smith new, in the order they are asked:
// the few every first workspace needs, then the ones behind "more options".
func NewQuestions() []NewQuestion {
	var languages []NewOption
	languages = append(languages, NewOption{Label: "Decide later", Value: LanguageLater, Hint: "add projects in any language with smith add (default)"})
	for _, id := range []string{"node", "python", "go", "jvm", "dotnet", "rust"} {
		languages = append(languages, NewOption{Label: languagenames.Of(id), Value: id})
	}
	var layouts []NewOption
	for _, a := range architecture.Catalog() {
		hint := layoutHints[a.Name]
		if a.Name == DefaultArchitecture {
			hint += " (default)"
		}
		layouts = append(layouts, NewOption{Label: a.Name, Value: a.Name, Hint: hint})
	}
	isNode := func(a map[string]string) bool { return a[AnswerLanguage] == "node" }
	more := func(a map[string]string) bool { return a[AnswerMore] == MoreYes }
	hasLanguage := func(a map[string]string) bool { return a[AnswerLanguage] != "" && a[AnswerLanguage] != LanguageLater }
	saving := func(a map[string]string) bool { return hasLanguage(a) && a[AnswerSaveRecipe] == "yes" }
	return []NewQuestion{
		{Key: AnswerCI, Prompt: "Which CI should smith write pipelines for?", Default: "github", Flag: "ci", Options: []NewOption{
			{Label: "GitHub Actions", Value: "github", Hint: "default"},
			{Label: "Azure Pipelines", Value: "azure"},
			{Label: "GitLab CI", Value: "gitlab"},
			{Label: "CircleCI", Value: "circleci"},
		}},
		{Key: AnswerLanguage, Prompt: "What is the main language of this workspace?", Default: LanguageLater, Flag: "language", Options: languages},
		{Key: AnswerNodeStack, Prompt: "Which tests and lint for Node projects?", Default: DefaultNodeStack, Flag: "default-stack", Options: NodeStacks(), Applies: isNode},
		{Key: AnswerArchitecture, Prompt: "Which code layout should smith check?", Default: DefaultArchitecture, Flag: "architecture", Options: layouts},
		{Key: AnswerMore, Prompt: "Change anything else? Release branch and remote, GitHub Releases, the cache.", Default: "no", Flag: "", Options: []NewOption{
			{Label: "Keep the defaults", Value: "no", Hint: "you can change any of it later in smith.toml"},
			{Label: "More options", Value: MoreYes},
		}},
		{Key: AnswerReleaseBranch, Prompt: "Which branch do releases run from?", Default: DefaultReleaseBranch, Flag: "release-branch", Advanced: true, Applies: more, Check: nonEmpty("a branch name")},
		{Key: AnswerReleaseRemote, Prompt: "Which git remote receives the release tags?", Default: DefaultReleaseRemote, Flag: "release-remote", Advanced: true, Applies: more, Check: nonEmpty("a remote name")},
		{Key: AnswerGitHubReleases, Prompt: "Create a GitHub Release for every release tag?", Default: "no", Flag: "github-releases", Advanced: true, Applies: more, Options: []NewOption{
			{Label: "No", Value: "no", Hint: "default; tags only"},
			{Label: "Yes", Value: "yes", Hint: "needs GITHUB_TOKEN in CI"},
		}},
		{Key: AnswerCacheDir, Prompt: "Where should the task cache live (a folder inside the workspace)?", Default: DefaultCacheDir, Flag: "cache-dir", Advanced: true, Applies: more, Check: nonEmpty("a folder")},
		{Key: AnswerCacheMaxMB, Prompt: "How large may the cache grow, in MB (0 for no limit)?", Default: strconv.Itoa(DefaultCacheMaxMB), Flag: "cache-max-mb", Advanced: true, Applies: more, Check: func(v string) error {
			if n, err := strconv.Atoi(v); err != nil || n < 0 {
				return fmt.Errorf("%q is not a size in MB: use a whole number, 0 for no limit", v)
			}
			return nil
		}},
		{Key: AnswerSaveRecipe, Prompt: "Save these answers as a recipe to reuse on other workspaces?", Default: "no", AtEnd: true, Applies: hasLanguage, Options: []NewOption{
			{Label: "No", Value: "no", Hint: "default"},
			{Label: "Yes", Value: "yes", Hint: "writes recipes/<name>.recipe.toml: language, Node tests and lint, code layout"},
		}},
		{Key: AnswerRecipeName, Prompt: "Name of the recipe (Enter uses the workspace name)", Flag: "save-recipe", AtEnd: true, Applies: saving, Check: func(v string) error {
			if v == "" {
				return nil
			}
			return ValidateName("recipe", v)
		}},
	}
}

func nonEmpty(what string) func(string) error {
	return func(v string) error {
		if v == "" {
			return fmt.Errorf("type %s", what)
		}
		return nil
	}
}
