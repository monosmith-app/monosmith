package scaffold

import (
	"fmt"
	"strconv"
	"strings"
)

// WithAnswers returns opts with the answers to NewQuestions applied. A key
// that is missing or empty keeps its default, and an answer a question does
// not accept is an error that names the flag.
func WithAnswers(opts NewOptions, answers map[string]string) (NewOptions, error) {
	byKey := map[string]NewQuestion{}
	for _, q := range NewQuestions() {
		byKey[q.Key] = q
	}
	for key, value := range answers {
		q, ok := byKey[key]
		if !ok {
			return opts, fmt.Errorf("no question called %q", key)
		}
		if err := checkAnswer(q, value); err != nil {
			return opts, err
		}
	}
	get := func(key string) string {
		if v := strings.TrimSpace(answers[key]); v != "" {
			return v
		}
		return byKey[key].Default
	}
	if len(opts.Providers) == 0 {
		opts.Providers = strings.Split(get(AnswerCI), ",")
	}
	if lang := get(AnswerLanguage); lang != LanguageLater {
		opts.Language = lang
	}
	if stack := get(AnswerNodeStack); stack != DefaultNodeStack {
		opts.DefaultStacks = map[string]string{"node": stack}
	}
	opts.Architecture = get(AnswerArchitecture)
	opts.NoAgent = get(AnswerAgent) == "no"
	opts.NoDevcontainer = get(AnswerDevcontainer) == "no"
	opts.ReleaseBranch = get(AnswerReleaseBranch)
	opts.ReleaseRemote = get(AnswerReleaseRemote)
	opts.GitHubReleases = get(AnswerGitHubReleases) == "yes"
	opts.CacheDir = get(AnswerCacheDir)
	// 0 is "no limit" for the person; the options keep 0 for "not asked".
	if n, _ := strconv.Atoi(get(AnswerCacheMaxMB)); n == 0 {
		opts.CacheMaxMB = -1
	} else {
		opts.CacheMaxMB = n
	}
	return opts, nil
}

// checkAnswer says whether value is something the question accepts. The
// message names the flag, since that is how a script gives it.
func checkAnswer(q NewQuestion, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if q.Key == AnswerCI {
		for _, p := range strings.Split(value, ",") {
			if !hasOption(q.Options, p) {
				return fmt.Errorf("--ci %s: %q is not a CI smith writes pipelines for; use %s", value, p, optionValues(q.Options))
			}
		}
		return nil
	}
	if q.Options != nil && !hasOption(q.Options, value) {
		return fmt.Errorf("--%s %s: use one of %s", q.Flag, value, optionValues(q.Options))
	}
	if q.Check != nil {
		if err := q.Check(value); err != nil {
			return fmt.Errorf("--%s: %w", q.Flag, err)
		}
	}
	return nil
}

func hasOption(options []NewOption, value string) bool {
	for _, o := range options {
		if o.Value == value {
			return true
		}
	}
	return false
}

func optionValues(options []NewOption) string {
	values := make([]string, len(options))
	for i, o := range options {
		values[i] = o.Value
	}
	return strings.Join(values, ", ")
}
