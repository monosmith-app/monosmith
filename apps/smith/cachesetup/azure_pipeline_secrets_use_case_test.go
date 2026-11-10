package cachesetup

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/cachekey"
)

const flowSecret = "flow-secret-9f8e7d6c5b4a"

// pipelineWorld fakes git and az for the Azure Pipelines step and records
// every program run, so a test can look for the secret in all of them.
type pipelineWorld struct {
	t          *testing.T
	remote     string
	remoteErr  error
	listJSON   string
	listErrors []string
	definition string
	putCount   int
	ran        []string
	stdins     []string
}

func (w *pipelineWorld) run(stdin string, argv ...string) (string, string, error) {
	line := CommandLine(argv)
	w.ran, w.stdins = append(w.ran, line), append(w.stdins, stdin)
	switch {
	case line == "git remote get-url origin":
		return w.remote, "", w.remoteErr
	case strings.HasPrefix(line, "az pipelines list"):
		if len(w.listErrors) > 0 {
			e := w.listErrors[0]
			w.listErrors = w.listErrors[1:]
			return "", e, errors.New("exit 1")
		}
		return w.listJSON, "", nil
	case strings.HasPrefix(line, "az login"):
		return "", "", nil
	case strings.Contains(line, "--http-method get"):
		return w.definition, "", nil
	case strings.Contains(line, "--http-method put"):
		body, _ := os.ReadFile(argv[len(argv)-1])
		w.putCount++
		w.definition = string(body)
		return "", "", nil
	}
	w.t.Fatalf("unexpected %q", line)
	return "", "", nil
}

// flowFor is a flow whose dialogue answers from input, one line each.
func flowFor(w *pipelineWorld, input string, out *bytes.Buffer) flow {
	p := Ports{
		Run:      w.run,
		LookPath: func(string) (string, error) { return "x", nil },
		Dialogue: Terminal{In: bufio.NewReader(strings.NewReader(input)), Out: out},
	}
	return flow{Ports: p, facts: Facts{Providers: []string{"azure"}}, azure: &azureState{}}
}

func newWorld(t *testing.T, definition string) *pipelineWorld {
	return &pipelineWorld{t: t, remote: "https://acme@dev.azure.com/acme/Shop/_git/shop\n", definition: definition,
		listJSON: `[{"id":12,"name":"ci","path":"\\"}]`}
}

// press is the person pressing the action's key and saying yes.
func press(t *testing.T, actions []cachekey.KeyAction, secret string) (asked string) {
	t.Helper()
	if len(actions) != 1 || actions[0].Key != 'a' {
		t.Fatalf("actions %+v", actions)
	}
	actions[0].Run(secret, func(command string) bool { asked = command; return true })
	return asked
}

func assertNoSecret(t *testing.T, w *pipelineWorld, out *bytes.Buffer, asked string) {
	t.Helper()
	all := strings.Join(append(append(w.ran, w.stdins...), out.String(), asked), "\n")
	if strings.Contains(all, flowSecret) {
		t.Errorf("the secret leaked:\n%s", all)
	}
}

func TestTheSecretIsSetAsAPipelineVariableAfterAYesAndNeverLeaks(t *testing.T) {
	w, out := newWorld(t, `{"id":12,"variables":{}}`), &bytes.Buffer{}
	f := flowFor(w, "1\n", out) // yes to the offer
	asked := press(t, f.azureActions("SMITH_CACHE_KEY"), flowSecret)
	for _, want := range []string{"organisation: acme", "Project: Shop", "Pipeline: ci", "SMITH_CACHE_KEY will be created", "not shown"} {
		if !strings.Contains(asked, want) {
			t.Errorf("the confirmation lacks %q:\n%s", want, asked)
		}
	}
	if w.putCount != 1 || !strings.Contains(w.definition, `"isSecret":true`) || !strings.Contains(w.definition, flowSecret) {
		t.Fatalf("puts %d, definition %s", w.putCount, w.definition)
	}
	if !strings.Contains(out.String(), "smith ci generate already maps it") {
		t.Errorf("no word on env: mapping:\n%s", out)
	}
	if lines := f.azureDoneLines(); len(lines) != 1 || !strings.Contains(lines[0], "SMITH_CACHE_KEY") {
		t.Errorf("done lines %v", lines)
	}
	assertNoSecret(t, w, out, asked)
}

func TestNothingIsWrittenWhenThePersonSaysNoToTheConfirmation(t *testing.T) {
	w, out := newWorld(t, `{"id":12}`), &bytes.Buffer{}
	f := flowFor(w, "1\n", out)
	actions := f.azureActions("SMITH_CACHE_KEY")
	actions[0].Run(flowSecret, func(string) bool { return false })
	if w.putCount != 0 || len(f.azureDoneLines()) != 0 {
		t.Fatalf("puts %d", w.putCount)
	}
	for _, line := range w.ran {
		if strings.Contains(line, "--http-method put") {
			t.Fatal("a write ran after a no")
		}
	}
}

func TestADeclinedOfferRunsNothing(t *testing.T) {
	w, out := newWorld(t, `{}`), &bytes.Buffer{}
	f := flowFor(w, "2\n", out)
	if a := f.azureActions("SMITH_CACHE_KEY"); a != nil || len(w.ran) != 0 {
		t.Fatalf("actions %v, ran %v", a, w.ran)
	}
}

func TestAnExistingVariableIsReplacedOnlyWhenAsked(t *testing.T) {
	def := `{"id":12,"variables":{"SMITH_CACHE_KEY":{"isSecret":true}}}`
	// yes to the offer, then Replace it (1).
	w, out := newWorld(t, def), &bytes.Buffer{}
	f := flowFor(w, "1\n1\n", out)
	asked := press(t, f.azureActions("SMITH_CACHE_KEY"), flowSecret)
	if !strings.Contains(asked, "will be replaced") || w.putCount != 1 {
		t.Fatalf("asked %q, puts %d", asked, w.putCount)
	}
	assertNoSecret(t, w, out, asked)
	// yes to the offer, then Leave it (2): no action, no write.
	w, out = newWorld(t, def), &bytes.Buffer{}
	f = flowFor(w, "1\n2\n", out)
	if a := f.azureActions("SMITH_CACHE_KEY"); a != nil || w.putCount != 0 || !strings.Contains(out.String(), "stays as it is") {
		t.Fatalf("actions %v, puts %d\n%s", a, w.putCount, out)
	}
}

func TestSeveralPipelinesAreAskedAndTheChosenOneGetsTheVariable(t *testing.T) {
	w, out := newWorld(t, `{"id":13,"variables":{}}`), &bytes.Buffer{}
	w.listJSON = `[{"id":12,"name":"ci","path":"\\"},{"id":13,"name":"nightly","path":"\\ops"}]`
	f := flowFor(w, "1\n2\n", out) // yes; the second pipeline
	asked := press(t, f.azureActions("AZURE_STORAGE_SAS_TOKEN"), flowSecret)
	if !strings.Contains(asked, "Pipeline: nightly") {
		t.Fatalf("asked %q", asked)
	}
	var put string
	for _, l := range w.ran {
		if strings.Contains(l, "--http-method put") {
			put = l
		}
	}
	if !strings.Contains(put, "definitionId=13") {
		t.Errorf("wrong pipeline: %s", put)
	}
}

func TestMissingPiecesFallBackToTheManualStepsWithTheirName(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*pipelineWorld, *flow)
		input string
		want  string
	}{
		{"az", func(_ *pipelineWorld, f *flow) {
			f.LookPath = func(string) (string, error) { return "", errors.New("no") }
		}, "1\n", "`az` is not installed"},
		{"git", func(w *pipelineWorld, _ *flow) { w.remoteErr = errors.New("no remote") }, "1\n", "could not read the git remote"},
		{"not azure", func(w *pipelineWorld, _ *flow) { w.remote = "git@github.com:acme/shop.git" }, "1\n", "not an Azure DevOps address"},
		{"extension", func(w *pipelineWorld, _ *flow) {
			w.listErrors = []string{"ERROR: 'pipelines' is misspelled or not recognized by the system."}
		}, "1\n", "az extension add --name azure-devops"},
		{"sign in declined", func(w *pipelineWorld, _ *flow) { w.listErrors = []string{"Please run 'az login'"} }, "1\n2\n", "Nobody is signed in to az"},
		{"no pipeline", func(w *pipelineWorld, _ *flow) { w.listJSON = `[]` }, "1\n", "has no pipeline for the repository shop"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w, out := newWorld(t, `{}`), &bytes.Buffer{}
			f := flowFor(w, c.input, out)
			c.setup(w, &f)
			if a := f.azureActions("SMITH_CACHE_KEY"); a != nil {
				t.Fatalf("actions %v", a)
			}
			if !strings.Contains(out.String(), c.want) {
				t.Errorf("output lacks %q:\n%s", c.want, out)
			}
		})
	}
}

func TestSignInIsOfferedAndRunOnlyAfterAYes(t *testing.T) {
	w, out := newWorld(t, `{"id":12}`), &bytes.Buffer{}
	w.listErrors = []string{"Please run 'az login' to setup account."}
	f := flowFor(w, "1\n1\n", out) // yes to the offer, yes to the sign-in
	if a := f.azureActions("SMITH_CACHE_KEY"); len(a) != 1 {
		t.Fatalf("actions %v\n%s", a, out)
	}
	if !strings.Contains(strings.Join(w.ran, "\n"), "az login") {
		t.Errorf("no sign-in ran: %v", w.ran)
	}
}

func TestOtherProvidersGetNoAzureAction(t *testing.T) {
	w, out := newWorld(t, `{}`), &bytes.Buffer{}
	f := flowFor(w, "", out)
	f.facts.Providers = []string{"github"}
	if a := f.azureActions("SMITH_CACHE_KEY"); a != nil || len(w.ran) != 0 {
		t.Fatalf("%v %v", a, w.ran)
	}
}
