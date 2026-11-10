package cachesetup

import (
	"bufio"
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
)

// goBack is the script item that makes the dialogue ask to go back.
type goBack struct{}

// scriptEnd is what the scripted dialogue panics with when the script has no answer left.
type scriptEnd struct{}

type asked struct {
	op, prompt, def string
	canBack         bool
	earlier         *ConfirmAnswer
}

// scripted is a Dialogue that answers from a list: a string for Choose, Text
// and Show, a bool for Confirm, goBack{} to go back.
type scripted struct {
	script  []any
	asked   []asked
	said    []string
	secrets int
	canBack bool
}

func (s *scripted) next(a asked) any {
	a.canBack = s.canBack
	s.asked = append(s.asked, a)
	if len(s.script) == 0 {
		panic(scriptEnd{})
	}
	v := s.script[0]
	s.script = s.script[1:]
	if _, ok := v.(goBack); ok {
		panic(BackRequest{})
	}
	return v
}

func (s *scripted) AllowBack(can bool) { s.canBack = can }
func (s *scripted) Say(text string)    { s.said = append(s.said, text) }
func (s *scripted) Choose(prompt string, _ []Option, def string) string {
	return s.next(asked{op: StepChoose, prompt: prompt, def: def}).(string)
}
func (s *scripted) Text(prompt, def string) string {
	return s.next(asked{op: StepText, prompt: prompt, def: def}).(string)
}
func (s *scripted) Confirm(c Confirmation) (bool, bool) {
	return s.next(asked{op: StepConfirm, prompt: c.Prompt, earlier: c.Earlier}).(bool), false
}
func (s *scripted) Working(string) func() { return func() {} }
func (s *scripted) Show(title string, _ []string) {
	s.next(asked{op: StepShow, prompt: title})
}
func (s *scripted) Secret(string, string, func() (string, error), []cachekey.KeyAction) error {
	s.secrets++
	return nil
}

// walk runs the setup against the script; the script ending stops it.
func walk(facts Facts, d *scripted, az *fakeAz, written *[]string) (remote string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(scriptEnd); !ok {
				panic(r)
			}
		}
	}()
	return Run(facts, Ports{
		Dialogue: d, Run: az.run,
		LookPath: func(string) (string, error) { return "x", nil },
		Now:      func() time.Time { return time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC) },
		WriteRemote: func(r string, _ bool) (string, error) {
			*written = append(*written, r)
			return "", nil
		},
	})
}

func TestTheStepsAreListedAsQuestionsOrActions(t *testing.T) {
	want := map[string]bool{
		StepChoose: true, StepText: true, StepConfirm: true, StepShow: true,
		StepRun: false, StepWrite: false, StepSecret: false,
	}
	if got := Steps(); !reflect.DeepEqual(got, want) {
		t.Fatalf("steps %v, want %v", got, want)
	}
	// The result is a copy: changing it does not change the policy.
	Steps()[StepRun] = true
	if Steps()[StepRun] {
		t.Fatal("Steps returned the policy itself")
	}
}

func TestBackAsksTheEarlierQuestionAgainWithTheEarlierAnswer(t *testing.T) {
	d := &scripted{script: []any{
		KindHTTPS, "https://a.dev/x", // answer A
		goBack{},                       // at "Write it?": back to the address
		"https://b.dev/y", true, false, // answer B, write, no key
		"", "", // the last page(s)
	}}
	var written []string
	remote, err := walk(Facts{WorkspaceName: "shop"}, d, &fakeAz{}, &written)
	if err != nil || remote != "https://b.dev/y" {
		t.Fatalf("%q %v", remote, err)
	}
	if len(written) != 1 || written[0] != "https://b.dev/y" {
		t.Fatalf("written %v: only the final answer is written, once", written)
	}
	var addresses []asked
	for _, a := range d.asked {
		if a.op == StepText {
			addresses = append(addresses, a)
		}
	}
	if len(addresses) != 2 || addresses[0].def != "" || addresses[1].def != "https://a.dev/x" {
		t.Fatalf("the address must come again with answer A as its default: %+v", addresses)
	}
}

func TestBackToTheKindOfStorageKeepsTheEarlierChoiceMarked(t *testing.T) {
	d := &scripted{script: []any{KindHTTPS, goBack{}, KindS3}}
	var written []string
	walk(Facts{}, d, &fakeAz{}, &written)
	var kinds []asked
	for _, a := range d.asked {
		if a.op == StepChoose && strings.HasPrefix(a.prompt, "What kind") {
			kinds = append(kinds, a)
		}
	}
	if len(kinds) != 2 || kinds[0].def != KindAzure || kinds[1].def != KindHTTPS {
		t.Fatalf("kind asked %+v", kinds)
	}
	if kinds[0].canBack || !d.canBack {
		t.Fatalf("the first question has no way back and later ones have: %+v", kinds)
	}
}

func TestBackOnTheFirstQuestionAsksItAgain(t *testing.T) {
	d := &scripted{script: []any{goBack{}}}
	var written []string
	walk(Facts{}, d, &fakeAz{}, &written)
	if len(d.asked) != 2 || d.asked[0].prompt != d.asked[1].prompt {
		t.Fatalf("asked %+v", d.asked)
	}
	if !strings.Contains(strings.Join(d.said, "\n"), "first question") {
		t.Fatalf("said %v", d.said)
	}
}

func TestBackOnAConfirmationMarksTheEarlierAnswer(t *testing.T) {
	d := &scripted{script: []any{KindHTTPS, "https://a.dev/x", false, goBack{}, false}}
	var written []string
	walk(Facts{}, d, &fakeAz{}, &written)
	var writes []asked
	for _, a := range d.asked {
		if a.op == StepConfirm && a.prompt == "Write it?" {
			writes = append(writes, a)
		}
	}
	if len(writes) != 2 || writes[0].earlier != nil || writes[1].earlier == nil || writes[1].earlier.Yes {
		t.Fatalf("the answer no must come back marked: %+v", writes)
	}
}

func TestBackOverAProgramThatRanIsRefusedAndNothingRunsTwice(t *testing.T) {
	az := &fakeAz{}
	d := &scripted{script: []any{
		KindAzure,
		true, "sub-1", // list subscriptions (ran), pick one
		true,     // run the account list (ran)
		goBack{}, // at the account choice: the listing already ran
		// the account choice comes again; the script then ends
	}}
	var written []string
	walk(Facts{WorkspaceName: "shop"}, d, az, &written)
	if !strings.Contains(strings.Join(d.said, "\n"), "That step already ran; nothing was undone.") {
		t.Fatalf("the refusal was not said: %v", d.said)
	}
	accounts := 0
	for _, a := range d.asked {
		if a.prompt == "Which storage account?" {
			accounts++
		}
	}
	if accounts != 2 {
		t.Fatalf("the screen after the step must come again, asked %d times: %+v", accounts, d.asked)
	}
	if !reflect.DeepEqual(az.ran, []string{CommandLine(AzureSubscriptions().Argv), CommandLine(AzureAccounts("sub-1").Argv)}) {
		t.Fatalf("a listing ran twice or not at all: %v", az.ran)
	}
}

func TestBackOverTheWrittenSettingIsRefusedAndItIsWrittenOnce(t *testing.T) {
	d := &scripted{script: []any{
		KindHTTPS, "https://a.dev/x", true, // written
		goBack{}, // at the key question: the file was written
		false,    // no key
		"", "",   // the pages
	}}
	var written []string
	remote, err := walk(Facts{}, d, &fakeAz{}, &written)
	if err != nil || remote != "https://a.dev/x" || len(written) != 1 {
		t.Fatalf("%q %v written %v", remote, err, written)
	}
	if !strings.Contains(strings.Join(d.said, "\n"), "nothing was undone") {
		t.Fatalf("said %v", d.said)
	}
}

func TestASecretIsMadeOnceAndGoingBackAfterItNeverMakesItAgain(t *testing.T) {
	d := &scripted{script: []any{
		KindHTTPS, "https://a.dev/x", true, // written
		true,     // make the key: the secret screen opens once
		"",       // first page (the provider's steps)
		goBack{}, // second page back to the first: allowed, no action between
		"", "",   // both pages again
	}}
	var written []string
	_, err := walk(Facts{Providers: []string{"github"}}, d, &fakeAz{}, &written)
	if err != nil {
		t.Fatal(err)
	}
	if d.secrets != 1 || len(written) != 1 {
		t.Fatalf("the secret screen opened %d times and the file was written %d times: both must be once", d.secrets, len(written))
	}
	pages := 0
	for _, a := range d.asked {
		if a.op == StepShow {
			pages++
		}
	}
	if pages != 4 {
		t.Fatalf("the first page must come again: %d pages shown", pages)
	}
}

func TestBackOverASecretScreenIsRefused(t *testing.T) {
	d := &scripted{script: []any{
		KindHTTPS, "https://a.dev/x", false, // not written
		true,     // make the key
		goBack{}, // at the first page: the key was made
		"", "",
	}}
	var written []string
	walk(Facts{}, d, &fakeAz{}, &written)
	if d.secrets != 1 || !strings.Contains(strings.Join(d.said, "\n"), "nothing was undone") {
		t.Fatalf("secrets %d said %v", d.secrets, d.said)
	}
}

func TestTheTerminalGoesBackWithB(t *testing.T) {
	var out bytes.Buffer
	// https; address; b at "Write it?"; the address again (Enter keeps it); write; no key.
	input := "3\nhttps://a.dev/x\nb\n\n1\n2\n"
	var written []string
	p := ports(input, &fakeAz{}, &out, &written)
	p.Dialogue = &Terminal{In: bufio.NewReader(strings.NewReader(input)), Out: &out}
	remote, err := Run(Facts{WorkspaceName: "shop"}, p)
	if err != nil || remote != "https://a.dev/x" || len(written) != 1 {
		t.Fatalf("%q %v written %v\n%s", remote, err, written, &out)
	}
	if !strings.Contains(out.String(), "b) Back") || !strings.Contains(out.String(), "[https://a.dev/x]") {
		t.Fatalf("the terminal must offer b and show the earlier address:\n%s", &out)
	}
	// The write question twice and the key question once; never the first question.
	if strings.Count(out.String(), "b) Back") != 3 || strings.Index(out.String(), "b) Back") < strings.Index(out.String(), "Address of the cache server") {
		t.Fatalf("b is offered after the first question only:\n%s", &out)
	}
}
