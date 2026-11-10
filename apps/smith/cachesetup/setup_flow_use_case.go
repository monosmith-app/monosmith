package cachesetup

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
)

// Facts are what the setup knows of the workspace.
type Facts struct {
	// WorkspaceName names the workspace; its folder's name when smith.toml has none.
	WorkspaceName string
	// Remote is `[cache] remote` today, "" when none.
	Remote string
	// Providers are `[ci] providers`.
	Providers []string
}

// Ports are the setup's connections to the world; tests replace every one.
type Ports struct {
	// In and Out are the terminal.
	In  *bufio.Reader
	Out io.Writer
	// Run starts a program; LookPath finds one on PATH.
	Run      Runner
	LookPath func(string) (string, error)
	// Now is the clock.
	Now func() time.Time
	// WriteRemote writes `[cache] remote`; it is called only after a yes.
	WriteRemote func(remote string) error
	// Secret copies a secret and offers the choices after it (see
	// cachekey.Session). produce is nil for a key made by smith.
	Secret func(noun string, produce func() (string, error), actions []cachekey.KeyAction) error
}

const byHand = "__by_hand__"

type flow struct {
	Ports
	facts Facts
}

func (f flow) say(format string, args ...any) { fmt.Fprintf(f.Out, format+"\n", args...) }

func (f flow) yesNo(prompt, yes, yesHint, no, noHint string) bool {
	return choiceprompt.Ask(prompt, []choiceprompt.Option{{Label: yes, Value: "yes", Hint: yesHint}, {Label: no, Value: "no", Hint: noHint}}, "yes", f.In, f.Out) == "yes"
}

func (f flow) text(prompt string) string {
	fmt.Fprintf(f.Out, "\n%s: ", prompt)
	line, _ := f.In.ReadString('\n')
	return strings.TrimSpace(line)
}

// list shows a read-only command, runs it when the person agrees, and returns
// what it lists. ok is false when nothing could be listed and the person must
// type the answer.
func (f flow) list(l Listing) (items []Item, ok bool) {
	program := l.Argv[0]
	if _, err := f.LookPath(program); err != nil {
		f.say("\n`%s` is not installed here, so smith cannot list this for you. You can type it by hand.", program)
		return nil, false
	}
	for attempt := 0; attempt < 2; attempt++ {
		f.say("\nsmith will run this command. It only reads:\n  %s", l.Line())
		if !f.yesNo("Run it?", "Run it", "list the choices from "+program, "Type it by hand", "run nothing") {
			return nil, false
		}
		stdout, stderr, err := f.Run("", l.Argv...)
		if err == nil {
			items, perr := l.Parse(stdout)
			if perr != nil {
				f.say("%v", perr)
				return nil, false
			}
			return items, true
		}
		login, needs := NeedsSignIn(program, stderr)
		if !needs || attempt == 1 {
			f.say("`%s` failed: %s", program, strings.TrimSpace(stderr))
			return nil, false
		}
		f.say("\nNobody is signed in to %s.", program)
		if !f.yesNo("Sign in now?", "Sign in", "run `"+CommandLine(login)+"`, which opens your browser or asks you", "Type it by hand", "run nothing") {
			return nil, false
		}
		if _, e, err := f.Run("", login...); err != nil {
			f.say("`%s` failed: %s", CommandLine(login), strings.TrimSpace(e))
			return nil, false
		}
	}
	return nil, false
}

// choose asks for one item of a list, with typing by hand as the last choice.
func (f flow) choose(prompt string, items []Item) string {
	options := make([]choiceprompt.Option, 0, len(items)+1)
	for _, it := range items {
		options = append(options, choiceprompt.Option{Label: it.Label, Value: it.Value, Hint: it.Hint})
	}
	options = append(options, choiceprompt.Option{Label: "Type it by hand", Value: byHand, Hint: "none of these"})
	return choiceprompt.Ask(prompt, options, options[0].Value, f.In, f.Out)
}

// pick lists with a command and asks which one; it falls back to typing.
func (f flow) pick(listing Listing, prompt, typed string) string {
	items, ok := f.list(listing)
	if ok && len(items) > 0 {
		if v := f.choose(prompt, items); v != byHand {
			return v
		}
	} else if ok {
		f.say("Nothing was listed.")
	}
	return f.text(typed)
}

// baseFolder walks the folders of a container or bucket so the person picks a
// base folder such as "smith". folders makes the listing for a prefix.
func (f flow) baseFolder(folders func(prefix string) Listing) string {
	prefix := ""
	for {
		items, ok := f.list(folders(prefix))
		if !ok {
			return strings.Trim(f.text("Base folder (for example smith; empty for the top)"), "/")
		}
		here := "Use the top as the base folder"
		if prefix != "" {
			here = "Use " + prefix + " as the base folder"
		}
		options := []choiceprompt.Option{{Label: here, Value: "__use__", Hint: "the workspace's sub-folder is proposed next"}}
		for _, it := range items {
			options = append(options, choiceprompt.Option{Label: "Open " + it.Label + "/", Value: it.Value, Hint: "look inside it"})
		}
		options = append(options, choiceprompt.Option{Label: "Name a new base folder", Value: "__new__", Hint: "it appears at the first write"})
		switch v := choiceprompt.Ask("Which base folder?", options, "__use__", f.In, f.Out); v {
		case "__use__":
			return prefix
		case "__new__":
			return JoinPrefix(prefix, f.text("Name of the new folder"))
		default:
			prefix = v
		}
	}
}

// workspaceFolder proposes a sub-folder named after the workspace.
func (f flow) workspaceFolder(base string) string {
	seg := JoinPrefix(base, WorkspaceSegment(f.facts.WorkspaceName))
	f.say("\nThe cache key of a task does not include the workspace. A prefix for each workspace keeps")
	f.say("workspaces apart and lets you delete one workspace's cache by deleting its prefix.")
	f.say("Workspaces that share code can share one prefix.")
	options := []choiceprompt.Option{
		{Label: "Use " + seg, Value: seg, Hint: "a folder for this workspace (recommended)"},
		{Label: "Use " + orTop(base), Value: base, Hint: "shared with every workspace that uses it"},
		{Label: "Type another", Value: byHand, Hint: "any prefix"},
	}
	v := choiceprompt.Ask("Which prefix for this workspace?", options, seg, f.In, f.Out)
	if v == byHand {
		return strings.Trim(f.text("Prefix"), "/")
	}
	return v
}

func orTop(s string) string {
	if s == "" {
		return "the top"
	}
	return s
}

// Run walks the whole setup: where, the setting, the key, the Azure SAS
// token, and the steps of each CI provider.
//
// Parameters: facts are what smith knows of the workspace; ports connect it to the world.
// Returns: the remote the person chose, "" when none; an error when the setup could not go on.
func Run(facts Facts, ports Ports) (remote string, err error) {
	f := flow{ports, facts}
	f.say("Remote cache setup. Every decision is yours: smith shows a command before it runs and writes nothing until you say yes.")
	if facts.Remote != "" {
		f.say("[cache] remote is %s today.", facts.Remote)
	}
	kind := choiceprompt.Ask("What kind of remote storage? (Google Cloud Storage is not supported.)", []choiceprompt.Option{
		{Label: "Azure Blob", Value: KindAzure, Hint: "a container in an Azure storage account"},
		{Label: "S3", Value: KindS3, Hint: "a bucket in Amazon S3 or an S3-compatible store"},
		{Label: "HTTPS server", Value: KindHTTPS, Hint: "any server that takes PUT, GET and HEAD"},
	}, KindAzure, f.In, f.Out)
	var sas *sasTarget
	switch kind {
	case KindAzure:
		sub := f.pick(AzureSubscriptions(), "Which Azure subscription?", "Subscription id or name (empty for the current one)")
		account := f.pick(AzureAccounts(sub), "Which storage account?", "Storage account name")
		container := f.pick(AzureContainers(sub, account), "Which container?", "Container name")
		f.say("\nBlob storage has no real folders: a prefix appears at the first write. (An account with a hierarchical namespace has real ones.)")
		base := f.baseFolder(func(prefix string) Listing { return AzureFolders(sub, account, container, prefix) })
		remote = AzureURL(account, container, f.workspaceFolder(base))
		sas = &sasTarget{sub, account, container}
	case KindS3:
		bucket := f.pick(S3Buckets(), "Which bucket?", "Bucket name")
		base := f.baseFolder(func(prefix string) Listing { return S3Folders(bucket, prefix) })
		remote = S3URL(bucket, f.workspaceFolder(base))
	default:
		for {
			remote = f.text("Address of the cache server (https://...)")
			msg := ValidateHTTPS(remote)
			if msg == "" {
				break
			}
			f.say("%s", msg)
		}
	}

	f.say("\nThis changes smith.toml:\n\n  [cache]\n  remote = %q\n", remote)
	if f.yesNo("Write it?", "Yes, write it", "only the setting changes; comments stay", "No, change nothing", "set it yourself with `smith config set cache.remote "+remote+"`") {
		if err := f.WriteRemote(remote); err != nil {
			return remote, err
		}
		f.say("Written. Another workspace uses another prefix, so its cache stays apart.")
	}

	if f.yesNo("\nMake the shared key (SMITH_CACHE_KEY) now?", "Yes, make and copy a key", "it goes to the clipboard and is never shown; the key is not stored anywhere", "No, I have one", "`smith cache key` makes one later") {
		if err := f.Secret("Key", nil, f.keyActions(facts.Providers, remote)); err != nil {
			return remote, err
		}
	}
	if sas != nil {
		f.sasToken(*sas)
	}
	f.guides(remote)
	return remote, nil
}

// keyActions are the choices that send the key to a provider's secret store.
func (f flow) keyActions(providers []string, remote string) []cachekey.KeyAction {
	var actions []cachekey.KeyAction
	for _, g := range ProviderGuides(providers, remote) {
		if g.KeyCommand == nil {
			continue
		}
		g := g
		actions = append(actions, cachekey.KeyAction{Key: map[string]rune{"github": 'h', "gitlab": 'l'}[g.Provider], Label: "send to " + g.Title + " (" + g.KeyCommand[0] + ")",
			Run: func(secret string, confirm func(string) bool) {
				if _, err := f.LookPath(g.KeyCommand[0]); err != nil {
					f.say("`%s` is not installed here. Use the steps below.", g.KeyCommand[0])
					return
				}
				if !confirm(CommandLine(g.KeyCommand) + "   (the key goes in on standard input)") {
					return
				}
				if _, stderr, err := f.Run(secret+"\n", g.KeyCommand...); err != nil {
					f.say("`%s` failed: %s", g.KeyCommand[0], strings.TrimSpace(stderr))
					return
				}
				f.say("The key is stored for %s.", g.Title)
			}})
	}
	return actions
}

type sasTarget struct{ subscription, account, container string }

// sasToken offers to make the container's SAS token with az and copy it.
func (f flow) sasToken(t sasTarget) {
	f.say("\nAzure Blob also needs AZURE_STORAGE_SAS_TOKEN: a token that lets smith reach the container. Signing in through az directly is not supported by the backend today.")
	if !f.yesNo("Make a SAS token now?", "Yes, make and copy one", "read, write, create and list permission, for this container only", "No, I will make one", "") {
		return
	}
	var options []choiceprompt.Option
	for _, c := range SASChoices() {
		options = append(options, choiceprompt.Option{Label: c.Label, Value: c.Value, Hint: c.Hint})
	}
	way := choiceprompt.Ask("How should the token be signed?", options, "login", f.In, f.Out)
	days := []choiceprompt.Option{{Label: "7 days", Value: "7"}, {Label: "30 days", Value: "30"}, {Label: "90 days", Value: "90"}, {Label: "365 days", Value: "365"}}
	if way == "login" {
		days = days[:1]
	}
	n := choiceprompt.Ask("How long should it last? You must renew it before it ends.", days, days[0].Value, f.In, f.Out)
	d, _ := strconv.Atoi(n)
	argv := GenerateSAS(t.subscription, t.account, t.container, way, f.Now().AddDate(0, 0, d))
	f.say("\nsmith will run this command. It makes a token and changes nothing in Azure:\n  %s", CommandLine(argv))
	if !f.yesNo("Run it?", "Run it", "the token is copied, never shown", "Skip", "run nothing") {
		return
	}
	produce := func() (string, error) {
		out, stderr, err := f.Run("", argv...)
		if err != nil {
			return "", fmt.Errorf("`az` failed: %s", strings.TrimSpace(stderr))
		}
		return strings.Trim(strings.TrimSpace(out), `"`), nil
	}
	if err := f.Secret("SAS token", produce, nil); err != nil {
		f.say("%v", err)
	}
}

// guides prints the steps of each CI provider in smith.toml.
func (f flow) guides(remote string) {
	for _, g := range ProviderGuides(f.facts.Providers, remote) {
		f.say("\n== %s ==", g.Title)
		for i, s := range g.Steps {
			f.say("%d. %s", i+1, s)
		}
		for _, s := range g.Secrets {
			f.say("   - %s", s)
		}
		if g.KeyCommand != nil {
			f.say("From a terminal: %s   (the value goes in on standard input)", CommandLine(g.KeyCommand))
		}
		if g.ByHand != "" {
			f.say("From a terminal: %s", g.ByHand)
		}
		for _, n := range g.Notes {
			f.say("Note: %s", n)
		}
	}
	if len(f.facts.Providers) == 0 {
		f.say("\nsmith.toml names no CI provider. Add one with `smith config set ci.providers github` to get its steps.")
	}
	f.say("\nOn a developer machine, set SMITH_CACHE_KEY (and the credential) as environment variables, or keep them in a password manager. `smith doctor` checks them.")
}
