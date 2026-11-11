package cachesetup

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachecredentials"
	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
	"github.com/monosmith-app/monosmith/libs/words/words"
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
	// Dialogue talks to the person: a terminal, or the screens of the shell.
	Dialogue
	// Run starts a program; LookPath finds one on PATH.
	Run Runner
	// Read starts a program that only reads (a listing); nil means Run.
	Read     Runner
	LookPath func(string) (string, error)
	// Now is the clock.
	Now func() time.Time
	// WriteRemote writes `[cache] remote`; it is called only after a yes. commit
	// is true when the person ticked the commit box. It returns a note to tell
	// the person (why nothing was committed, or what was) and an error.
	WriteRemote func(remote string, commit bool) (note string, err error)
}

const byHand = "__by_hand__"

type flow struct {
	Ports
	facts Facts
	// azure is what the setup learned of the Azure Pipelines pipeline.
	azure *azureState
}

func (f flow) say(format string, args ...any) { f.Say(fmt.Sprintf(format, args...)) }

func (f flow) yesNo(prompt, yes, yesHint, no, noHint string) bool {
	ok, _ := f.Confirm(Confirmation{Prompt: prompt, Yes: yes, YesHint: yesHint, No: no, NoHint: noHint})
	return ok
}

func (f flow) text(prompt string) string { return f.Text(prompt, "") }

// run starts a program while the person sees that smith is working.
func (f flow) run(what string, stdin string, argv ...string) (string, string, error) {
	done := f.Working(what)
	defer done()
	return f.Run(stdin, argv...)
}

// read starts a program that only reads while the person sees that smith is working.
func (f flow) read(what string, stdin string, argv ...string) (string, string, error) {
	done := f.Working(what)
	defer done()
	if f.Read == nil {
		return f.Run(stdin, argv...)
	}
	return f.Read(stdin, argv...)
}

// list shows a read-only command, runs it when the person agrees, and returns
// what it lists. ok is false when nothing could be listed and the person must
// type the answer.
func (f flow) list(l Listing) (items []Item, ok bool) {
	program := l.Argv[0]
	if _, err := f.LookPath(program); err != nil {
		f.say("`%s` is not installed here, so smith cannot list this for you. You can type it by hand.", program)
		return nil, false
	}
	for attempt := 0; attempt < 2; attempt++ {
		if ok, _ := f.Confirm(Confirmation{Text: "smith will run this command. It only reads:", Shown: l.Line(), Prompt: "Run it?", Yes: "Run it", YesHint: "list the choices from " + program, No: "Type it by hand", NoHint: "run nothing"}); !ok {
			return nil, false
		}
		stdout, stderr, err := f.read("Listing with "+program, "", l.Argv...)
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
		if ok, _ := f.Confirm(Confirmation{Text: "Nobody is signed in to " + program + ". Signing in opens your browser or asks you:", Shown: CommandLine(login), Prompt: "Sign in now?", Yes: "Sign in", YesHint: "run the command above", No: "Type it by hand", NoHint: "run nothing"}); !ok {
			return nil, false
		}
		if _, e, err := f.run("Signing in with "+program, "", login...); err != nil {
			f.say("`%s` failed: %s", CommandLine(login), strings.TrimSpace(e))
			return nil, false
		}
	}
	return nil, false
}

// choose asks for one item of a list, with typing by hand as the last choice.
func (f flow) choose(prompt string, items []Item) string {
	options := make([]Option, 0, len(items)+1)
	for _, it := range items {
		options = append(options, Option{Label: it.Label, Value: it.Value, Hint: it.Hint})
	}
	options = append(options, Option{Label: "Type it by hand", Value: byHand, Hint: "none of these"})
	return f.Choose(prompt, options, options[0].Value)
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
		options := []Option{{Label: here, Value: "__use__", Hint: "the workspace's sub-folder is proposed next"}}
		for _, it := range items {
			options = append(options, Option{Label: "Open " + it.Label + "/", Value: it.Value, Hint: "look inside it"})
		}
		options = append(options, Option{Label: "Name a new base folder", Value: "__new__", Hint: "it appears at the first write"})
		switch v := f.Choose("Which base folder?", options, "__use__"); v {
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
	f.say("The cache key of a task does not include the workspace. A prefix for each workspace keeps workspaces apart and lets you delete one workspace's cache by deleting its prefix. Workspaces that share code can share one prefix.")
	options := []Option{
		{Label: "Use " + seg, Value: seg, Hint: "a folder for this workspace (recommended)"},
		{Label: "Use " + orTop(base), Value: base, Hint: "shared with every workspace that uses it"},
		{Label: "Type another", Value: byHand, Hint: "any prefix"},
	}
	v := f.Choose("Which prefix for this workspace?", options, seg)
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
// The person may go back to the question before the one on screen (the
// dialogue panics with BackRequest). Run then walks the flow again, answers the
// steps before that question from what it recorded, and asks the question with
// the old answer chosen. A step that ran something (see Steps) is a point of
// no return: Run does not go back over it, and says so.
//
// Parameters: facts are what smith knows of the workspace; ports connect it to the world.
// Returns: the remote the person chose, "" when none; an error when the setup could not go on.
func Run(facts Facts, ports Ports) (remote string, err error) {
	h := &history{}
	for {
		var again bool
		if remote, again, err = walkOnce(facts, ports, h); !again {
			return remote, err
		}
	}
}

// walkOnce walks the flow from its start, answering from the history first. It
// returns again true when the person went back and the flow must start over.
func walkOnce(facts Facts, ports Ports, h *history) (remote string, again bool, err error) {
	az := &azureState{}
	s := &stepper{inner: ports.Dialogue, run: ports.Run, write: ports.WriteRemote, h: h, azure: az}
	p := ports
	p.Dialogue, p.Run, p.Read, p.WriteRemote = s, s.runProgram, s.readProgram, s.writeRemote
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if _, ok := r.(BackRequest); !ok {
			panic(r)
		}
		again = true
		switch h.goBack() {
		case refusedBack:
			ports.Say(words.Say("cachesetup.back.refused"))
		case firstQuestion:
			ports.Say(words.Say("cachesetup.back.first"))
		}
	}()
	remote, err = flow{Ports: p, facts: facts, azure: az}.walk()
	return remote, false, err
}

// walk is the setup from the first question to the last page.
func (f flow) walk() (remote string, err error) {
	facts := f.facts
	f.say("Remote cache setup. Every decision is yours: smith shows a command before it runs and writes nothing until you say yes.")
	if facts.Remote != "" {
		f.say("[cache] remote is %s today.", facts.Remote)
	}
	kind := f.Choose("What kind of remote storage? (Google Cloud Storage is not supported.)", []Option{
		{Label: "Azure Blob", Value: KindAzure, Hint: "a container in an Azure storage account"},
		{Label: "S3", Value: KindS3, Hint: "a bucket in Amazon S3 or an S3-compatible store"},
		{Label: "HTTPS server", Value: KindHTTPS, Hint: "any server that takes PUT, GET and HEAD"},
	}, KindAzure)
	var sas *sasTarget
	switch kind {
	case KindAzure:
		sub := f.pick(AzureSubscriptions(), "Which Azure subscription?", "Subscription id or name (empty for the current one)")
		account := f.pick(AzureAccounts(sub), "Which storage account?", "Storage account name")
		container := f.pick(AzureContainers(sub, account), "Which container?", "Container name")
		f.say("Blob storage has no real folders: a prefix appears at the first write. (An account with a hierarchical namespace has real ones.)")
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

	if ok, commit := f.Confirm(Confirmation{Text: "This changes smith.toml:", Shown: fmt.Sprintf("[cache]\nremote = %q", remote), Prompt: "Write it?", Yes: "Yes, write it", YesHint: "only the setting changes; comments stay", No: "No, change nothing", NoHint: "set it yourself with `smith config set cache.remote " + remote + "`", OfferCommit: true}); ok {
		note, err := f.WriteRemote(remote, commit)
		if err != nil {
			return remote, err
		}
		f.say("Written. Another workspace uses another prefix, so its cache stays apart.")
		if note != "" {
			f.say("%s", note)
		}
	}

	if f.yesNo("\nMake the shared key (SMITH_CACHE_KEY) now?", "Yes, make and copy a key", "it goes to the clipboard and is never shown; the key is not stored anywhere", "No, I have one", "`smith cache key` makes one later") {
		if err := f.Secret("Key", cachekey.KeyVariable, nil, f.keyActions(facts.Providers, remote)); err != nil {
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
				f.say("Sending the key to %s…", g.Title)
				if _, stderr, err := f.Run(secret+"\n", g.KeyCommand...); err != nil {
					f.say("`%s` failed: %s", g.KeyCommand[0], strings.TrimSpace(stderr))
					return
				}
				f.say("The key is stored for %s.", g.Title)
			}})
	}
	return append(actions, f.azureActions(cachecredentials.Key)...)
}

type sasTarget struct{ subscription, account, container string }

// sasToken offers to make the container's SAS token with az and copy it.
func (f flow) sasToken(t sasTarget) {
	f.say("Azure Blob also needs AZURE_STORAGE_SAS_TOKEN: a token that lets smith reach the container. Signing in through az directly is not supported by the backend today.")
	if !f.yesNo("Make a SAS token now?", "Yes, make and copy one", "read, write, delete, create and list permission, for this container only (delete lets smith cache clear --remote work)", "No, I will make one", "") {
		return
	}
	var options []Option
	for _, c := range SASChoices() {
		options = append(options, Option{Label: c.Label, Value: c.Value, Hint: c.Hint})
	}
	var way, n string
	for {
		way = f.Choose("How should the token be signed?", options, "login")
		days := []Option{{Label: "7 days", Value: "7"}, {Label: "30 days", Value: "30"}, {Label: "90 days", Value: "90"}, {Label: "365 days", Value: "365"}}
		prompt := words.Say("cachekey.sas.expiry.other")
		if way == "login" {
			days, prompt = days[:1], words.Say("cachekey.sas.expiry.login")
		}
		days = append(days, Option{Label: words.Say("cachekey.sas.back"), Value: "back", Hint: words.Say("cachekey.sas.back.hint")})
		if n = f.Choose(prompt, days, days[0].Value); n != "back" {
			break
		}
	}
	d, _ := strconv.Atoi(n)
	argv := GenerateSAS(t.subscription, t.account, t.container, way, f.Now().AddDate(0, 0, d))
	if ok, _ := f.Confirm(Confirmation{Text: "smith will run this command. It makes a token and changes nothing in Azure:", Shown: CommandLine(argv), Prompt: "Run it?", Yes: "Run it", YesHint: "the token is copied, never shown", No: "Skip", NoHint: "run nothing"}); !ok {
		return
	}
	produce := func() (string, error) {
		out, stderr, err := f.Run("", argv...)
		if err != nil {
			return "", fmt.Errorf("`az` failed: %s", strings.TrimSpace(stderr))
		}
		return strings.Trim(strings.TrimSpace(out), `"`), nil
	}
	if err := f.Secret("SAS token", cachekey.SASVariable, produce, f.azureActions(sasTokenName)); err != nil {
		f.say("%v", err)
	}
}

// guides shows the steps of each CI provider in smith.toml, one page each.
func (f flow) guides(remote string) {
	for _, g := range ProviderGuides(f.facts.Providers, remote) {
		var lines []string
		if g.Provider == "azure" {
			lines = append(lines, f.azureDoneLines()...)
		}
		for i, s := range g.Steps {
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, s))
		}
		for _, s := range g.Secrets {
			lines = append(lines, "   - "+s)
		}
		if g.KeyCommand != nil {
			lines = append(lines, "From a terminal: "+CommandLine(g.KeyCommand)+"   (the value goes in on standard input)")
		}
		if g.ByHand != "" {
			lines = append(lines, "From a terminal: "+g.ByHand)
		}
		for _, n := range g.Notes {
			lines = append(lines, "Note: "+n)
		}
		f.Show(g.Title, lines)
	}
	var last []string
	if len(f.facts.Providers) == 0 {
		last = append(last, "smith.toml names no CI provider. Add one with `smith config set ci.providers github` to get its steps.")
	}
	last = append(last, "On a developer machine, set SMITH_CACHE_KEY (and the credential) as environment variables, or keep them in a password manager. `smith doctor` checks them.")
	f.Show("Developer machines", last)
}
