package cachesetup

import (
	"errors"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/apps/smith/cachepipeline"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// sasTokenName is the variable that holds the Azure Blob SAS token.
const sasTokenName = "AZURE_STORAGE_SAS_TOKEN"

// azureState is what the setup learned about the Azure Pipelines pipeline. It
// is shared by every copy of the flow, so the pipeline is asked for once.
type azureState struct {
	// resolved is true after the offer was made and the pipeline looked for.
	resolved bool
	// target is the chosen pipeline, nil when there is none.
	target *cachepipeline.Target
	// variables are the names the pipeline holds, with true for a secret.
	variables map[string]bool
	// set lists the variables this setup created or replaced.
	set []string
}

// azureActions is the key action that sends a secret to the pipeline as a
// secret variable, or nothing when the person declined, az cannot do it, or
// the person chose to leave the existing variable.
//
// The first call makes the offer and looks for the pipeline; every question
// comes here, before the secret exists. The action itself runs when the person
// presses its key, asks one last time with the facts, and only then writes.
func (f flow) azureActions(name string) []cachekey.KeyAction {
	if !slices.Contains(f.facts.Providers, "azure") {
		return nil
	}
	if !f.azure.resolved {
		f.azure.resolved = true
		f.findPipeline()
	}
	t := f.azure.target
	if t == nil {
		return nil
	}
	replace := false
	if _, exists := f.azure.variables[name]; exists {
		ans := f.Choose(words.Say("cachepipeline.exists.prompt", "name", name, "pipeline", t.PipelineName), []Option{
			{Label: words.Say("cachepipeline.exists.replace"), Value: "replace", Hint: words.Say("cachepipeline.exists.replace.hint")},
			{Label: words.Say("cachepipeline.exists.leave"), Value: "leave", Hint: words.Say("cachepipeline.exists.leave.hint")},
		}, "leave")
		if ans != "replace" {
			f.Say(words.Say("cachepipeline.exists.left", "name", name, "pipeline", t.PipelineName))
			return nil
		}
		replace = true
	}
	return []cachekey.KeyAction{{Key: 'a', Label: words.Say("cachepipeline.action.label"), Run: func(secret string, confirm func(string) bool) {
		f.sendSecret(*t, name, secret, replace, confirm)
	}}}
}

// sendSecret asks with the facts, then sets the variable. The secret reaches
// only cachepipeline.Client, which keeps it off every command line.
func (f flow) sendSecret(t cachepipeline.Target, name, secret string, replace bool, confirm func(string) bool) {
	if _, now := f.azure.variables[name]; now {
		replace = true // set earlier in this setup: a second press replaces it
	}
	what := words.Say("cachepipeline.confirm.created")
	if replace {
		what = words.Say("cachepipeline.confirm.replaced")
	}
	if !confirm(words.Say("cachepipeline.confirm.summary", "organization", t.Organization, "project", t.Project, "pipeline", t.PipelineName, "name", name, "what", what)) {
		return
	}
	f.Say(words.Say("cachepipeline.working", "name", name, "pipeline", t.PipelineName))
	created, err := cachepipeline.Client{Run: cachepipeline.Runner(f.Run)}.SetSecret(t, name, secret, replace)
	switch {
	case errors.Is(err, cachepipeline.ErrExists):
		f.azure.variables[name] = true
		f.Say(words.Say("cachepipeline.exists.meanwhile", "name", name, "pipeline", t.PipelineName))
	case err != nil:
		f.Say(words.Say("cachepipeline.failed", "name", name, "error", err.Error()))
	default:
		f.azure.variables[name] = true
		if !slices.Contains(f.azure.set, name) {
			f.azure.set = append(f.azure.set, name)
		}
		key := "cachepipeline.done.replaced"
		if created {
			key = "cachepipeline.done.created"
		}
		f.Say(words.Say(key, "name", name, "pipeline", t.PipelineName))
	}
}

// findPipeline makes the offer, reads the git remote, lists the pipelines of
// the repository and, when several match, asks which one. It leaves
// f.azure.target nil, with a word of why, when it cannot go on.
func (f flow) findPipeline() {
	if ok, _ := f.Confirm(Confirmation{Text: words.Say("cachepipeline.offer.text"), Prompt: words.Say("cachepipeline.offer.prompt"),
		Yes: words.Say("cachepipeline.offer.yes"), YesHint: words.Say("cachepipeline.offer.yes.hint"),
		No: words.Say("cachepipeline.offer.no"), NoHint: words.Say("cachepipeline.offer.no.hint")}); !ok {
		return
	}
	if _, err := f.LookPath("az"); err != nil {
		f.Say(words.Say("cachepipeline.az.missing"))
		return
	}
	remote, _, err := f.run(words.Say("cachepipeline.git.working"), "", "git", "remote", "get-url", "origin")
	if err != nil || strings.TrimSpace(remote) == "" {
		f.Say(words.Say("cachepipeline.git.none"))
		return
	}
	repo, ok := cachepipeline.ParseAzureRemote(remote)
	if !ok {
		f.Say(words.Say("cachepipeline.git.other"))
		return
	}
	out, ok := f.azRead(cachepipeline.PipelinesArgv(repo))
	if !ok {
		return
	}
	pipelines, err := cachepipeline.ParsePipelines(out)
	if err != nil {
		f.Say(words.Say("cachepipeline.az.failed", "error", err.Error()))
		return
	}
	var chosen cachepipeline.Pipeline
	switch len(pipelines) {
	case 0:
		f.Say(words.Say("cachepipeline.pipelines.none", "project", repo.Project, "repository", repo.Repository))
		return
	case 1:
		chosen = pipelines[0]
	default:
		options := make([]Option, 0, len(pipelines)+1)
		for _, p := range pipelines {
			options = append(options, Option{Label: p.Name, Value: p.Name, Hint: p.Path})
		}
		options = append(options, Option{Label: words.Say("cachepipeline.pipelines.skip"), Value: byHand, Hint: words.Say("cachepipeline.pipelines.skip.hint")})
		answer := f.Choose(words.Say("cachepipeline.pipelines.prompt"), options, options[0].Value)
		i := slices.IndexFunc(pipelines, func(p cachepipeline.Pipeline) bool { return p.Name == answer })
		if i < 0 {
			return
		}
		chosen = pipelines[i]
	}
	t := cachepipeline.Target{Organization: repo.Organization, Project: repo.Project, PipelineID: chosen.ID, PipelineName: chosen.Name}
	client := cachepipeline.Client{Run: cachepipeline.Runner(f.Run)}
	done := f.Working(words.Say("cachepipeline.list.working", "command", "az devops invoke --area build --resource definitions"))
	vars, err := client.Variables(t)
	done()
	if err != nil {
		f.Say(words.Say("cachepipeline.az.failed", "error", err.Error()))
		return
	}
	f.azure.target, f.azure.variables = &t, vars
}

// azRead runs a read-only az command while the person sees the command. When
// nobody is signed in it offers `az login` and tries again once; it says why
// and returns false for a missing extension or any other failure.
func (f flow) azRead(argv []string) (string, bool) {
	for attempt := 0; ; attempt++ {
		stdout, stderr, err := f.run(words.Say("cachepipeline.list.working", "command", CommandLine(argv)), "", argv...)
		if err == nil {
			return stdout, true
		}
		switch cachepipeline.Classify(stderr) {
		case cachepipeline.FailureExtension:
			f.Say(words.Say("cachepipeline.extension.missing"))
			return "", false
		case cachepipeline.FailureSignIn:
			if attempt == 0 && f.signIn() {
				continue
			}
			return "", false
		}
		f.Say(words.Say("cachepipeline.az.failed", "error", strings.TrimSpace(stderr)))
		return "", false
	}
}

// signIn offers `az login` and runs it only after a yes.
func (f flow) signIn() bool {
	login := []string{"az", "login"}
	if ok, _ := f.Confirm(Confirmation{Text: words.Say("cachepipeline.signin.text"), Shown: CommandLine(login), Prompt: words.Say("cachepipeline.signin.prompt"),
		Yes: words.Say("cachepipeline.signin.yes"), YesHint: words.Say("cachepipeline.signin.yes.hint"),
		No: words.Say("cachepipeline.signin.no"), NoHint: words.Say("cachepipeline.signin.no.hint")}); !ok {
		return false
	}
	if _, stderr, err := f.run("az login", "", login...); err != nil {
		f.Say(words.Say("cachepipeline.signin.failed", "error", strings.TrimSpace(stderr)))
		return false
	}
	return true
}

// azureDoneLines tell what this setup set in the pipeline, for the top of the
// provider's page; nothing when nothing was set.
func (f flow) azureDoneLines() []string {
	if f.azure == nil || len(f.azure.set) == 0 || f.azure.target == nil {
		return nil
	}
	t := f.azure.target
	return []string{words.Say("cachepipeline.done.page", "pipeline", t.PipelineName, "organization", t.Organization, "project", t.Project, "names", strings.Join(f.azure.set, ", "))}
}
