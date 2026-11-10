package cachepipeline

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const secret = "s3cr3t-value-0123456789abcdef"

var target = Target{Organization: "acme", Project: "Shop App", PipelineID: 12, PipelineName: "ci"}

// fakeAz is az for the Build Definitions API: it keeps one definition, records
// every call, and checks what each call could leak.
type fakeAz struct {
	t          *testing.T
	definition string
	calls      [][]string
	files      []string
	modes      []os.FileMode
	bodies     []string
	// putErrors are the stderr texts of the next put calls; each one fails.
	putErrors []string
	panicPut  bool
}

func (f *fakeAz) run(stdin string, argv ...string) (string, string, error) {
	f.calls = append(f.calls, argv)
	if stdin != "" {
		f.t.Errorf("stdin is not used, got %q", stdin)
	}
	for _, a := range argv {
		if strings.Contains(a, secret) {
			f.t.Errorf("the secret is on a command line: %q", argv)
		}
	}
	for _, e := range os.Environ() {
		if strings.Contains(e, secret) {
			f.t.Errorf("the secret is in the environment")
		}
	}
	line := strings.Join(argv, " ")
	switch {
	case strings.Contains(line, "--http-method get"):
		return f.definition, "", nil
	case strings.Contains(line, "--http-method put"):
		file := argv[len(argv)-1]
		if argv[len(argv)-2] != "--in-file" {
			f.t.Fatalf("no --in-file in %q", argv)
		}
		info, err := os.Stat(file)
		if err != nil {
			f.t.Fatal(err)
		}
		body, _ := os.ReadFile(file)
		f.files, f.modes, f.bodies = append(f.files, file), append(f.modes, info.Mode().Perm()), append(f.bodies, string(body))
		if f.panicPut {
			panic("boom")
		}
		if len(f.putErrors) > 0 {
			e := f.putErrors[0]
			f.putErrors = f.putErrors[1:]
			return "", e, errors.New("exit 1")
		}
		f.definition = string(body)
		return "", "", nil
	}
	f.t.Fatalf("unexpected call %q", argv)
	return "", "", nil
}

const baseDefinition = `{"id":12,"revision":7,"name":"ci","process":{"type":2,"yamlFilename":"azure-pipelines.yml"},` +
	`"variables":{"Old":{"value":"keep","allowOverride":true},"SMITH_CACHE_KEY":{"isSecret":true,"allowOverride":true}}}`

func newClient(t *testing.T, def string) (Client, *fakeAz) {
	f := &fakeAz{t: t, definition: def}
	return Client{Run: f.run}, f
}

func decode(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func noFilesLeft(t *testing.T, f *fakeAz) {
	t.Helper()
	for _, file := range f.files {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("the request file %s is still there (%v)", file, err)
		}
		if _, err := os.Stat(filepath.Dir(file)); !os.IsNotExist(err) {
			t.Errorf("the request folder of %s is still there", file)
		}
	}
}

func TestANewSecretIsAddedAndTheRestOfTheDefinitionStays(t *testing.T) {
	c, f := newClient(t, baseDefinition)
	created, err := c.SetSecret(target, "AZURE_STORAGE_SAS_TOKEN", secret, false)
	if err != nil || !created {
		t.Fatalf("created %v, err %v", created, err)
	}
	got := decode(t, f.definition)
	vars := got["variables"].(map[string]any)
	sas := vars["AZURE_STORAGE_SAS_TOKEN"].(map[string]any)
	if sas["value"] != secret || sas["isSecret"] != true {
		t.Errorf("variable %v", sas)
	}
	if vars["Old"].(map[string]any)["value"] != "keep" || got["revision"] != float64(7) || got["process"].(map[string]any)["yamlFilename"] != "azure-pipelines.yml" {
		t.Errorf("the definition changed: %s", f.definition)
	}
	last := f.calls[len(f.calls)-1]
	for _, want := range []string{"devops", "invoke", "--area", "build", "--resource", "definitions", "project=Shop App", "definitionId=12", "https://dev.azure.com/acme", "--http-method", "put", "--in-file"} {
		if !strings.Contains(strings.Join(last, "\x00"), want) {
			t.Errorf("put call lacks %q: %q", want, last)
		}
	}
	if runtime.GOOS != "windows" && f.modes[0] != 0o600 {
		t.Errorf("request file mode %v, want 0600", f.modes[0])
	}
	noFilesLeft(t, f)
}

func TestAReplaceKeepsTheOtherSettingsAndNeedsPermission(t *testing.T) {
	c, f := newClient(t, baseDefinition)
	if _, err := c.SetSecret(target, "SMITH_CACHE_KEY", secret, false); !errors.Is(err, ErrExists) {
		t.Fatalf("want ErrExists, got %v", err)
	}
	if len(f.files) != 0 {
		t.Fatal("nothing may be written without permission to replace")
	}
	created, err := c.SetSecret(target, "SMITH_CACHE_KEY", secret, true)
	if err != nil || created {
		t.Fatalf("created %v, err %v", created, err)
	}
	v := decode(t, f.definition)["variables"].(map[string]any)["SMITH_CACHE_KEY"].(map[string]any)
	if v["value"] != secret || v["isSecret"] != true || v["allowOverride"] != true {
		t.Errorf("variable %v", v)
	}
	noFilesLeft(t, f)
}

func TestAConflictReadsTheDefinitionAgainOnceAndNeverLeaksTheValue(t *testing.T) {
	c, f := newClient(t, baseDefinition)
	f.putErrors = []string{"ERROR: 409 Conflict: the revision is old"}
	if _, err := c.SetSecret(target, "AZURE_STORAGE_SAS_TOKEN", secret, false); err != nil {
		t.Fatal(err)
	}
	gets := 0
	for _, call := range f.calls {
		if strings.Contains(strings.Join(call, " "), "--http-method get") {
			gets++
		}
	}
	if gets != 2 || len(f.files) != 2 {
		t.Errorf("reads %d, writes %d; want 2 and 2", gets, len(f.files))
	}
	noFilesLeft(t, f)

	c, f = newClient(t, baseDefinition)
	f.putErrors = []string{"revision conflict " + secret, "revision conflict " + secret}
	_, err := c.SetSecret(target, "AZURE_STORAGE_SAS_TOKEN", secret, false)
	if err == nil || strings.Contains(err.Error(), secret) || !strings.Contains(err.Error(), "***") {
		t.Fatalf("error %v", err)
	}
	if len(f.files) != 2 {
		t.Errorf("a second conflict must stop after one retry, wrote %d times", len(f.files))
	}
	noFilesLeft(t, f)
}

func TestAnOtherFailureIsNotRetriedAndTheFileIsGone(t *testing.T) {
	c, f := newClient(t, baseDefinition)
	f.putErrors = []string{"ERROR: TF400813 denied " + secret}
	_, err := c.SetSecret(target, "AZURE_STORAGE_SAS_TOKEN", secret, false)
	if err == nil || strings.Contains(err.Error(), secret) || len(f.files) != 1 {
		t.Fatalf("err %v, writes %d", err, len(f.files))
	}
	noFilesLeft(t, f)
}

func TestThePrivateFolderIsRemovedWhenTheRunnerPanics(t *testing.T) {
	c, f := newClient(t, baseDefinition)
	f.panicPut = true
	func() {
		defer func() { _ = recover() }()
		_, _ = c.SetSecret(target, "AZURE_STORAGE_SAS_TOKEN", secret, false)
	}()
	if len(f.files) != 1 {
		t.Fatal("the put was not reached")
	}
	noFilesLeft(t, f)
}

func TestNoLogLineOrErrorOfAReadHoldsTheValue(t *testing.T) {
	c, _ := newClient(t, "not json")
	if _, err := c.SetSecret(target, "X", secret, false); err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("%v", err)
	}
	c.Run = func(string, ...string) (string, string, error) { return "", "no sign in", errors.New("x") }
	if _, err := c.Variables(target); err == nil || !strings.Contains(err.Error(), "no sign in") {
		t.Fatalf("%v", err)
	}
}

func TestVariablesTellWhichNamesExistAndWhichAreSecret(t *testing.T) {
	c, _ := newClient(t, baseDefinition)
	names, err := c.Variables(target)
	if err != nil || len(names) != 2 || names["Old"] || !names["SMITH_CACHE_KEY"] {
		t.Fatalf("%v %v", names, err)
	}
	c, _ = newClient(t, `{"id":1}`)
	if names, err := c.Variables(target); err != nil || len(names) != 0 {
		t.Fatalf("%v %v", names, err)
	}
}

func TestPipelinesAreListedForTheRepository(t *testing.T) {
	line := strings.Join(PipelinesArgv(AzureRepo{Organization: "acme", Project: "Shop App", Repository: "shop"}), " ")
	for _, want := range []string{"az pipelines list", "--organization https://dev.azure.com/acme", "--project Shop App", "--repository shop", "--repository-type tfsgit"} {
		if !strings.Contains(line, want) {
			t.Errorf("%q lacks %q", line, want)
		}
	}
	p, err := ParsePipelines(`[{"id":3,"name":"ci","path":"\\"},{"id":4,"name":"nightly","path":"\\ops"}]`)
	if err != nil || len(p) != 2 || p[1].ID != 4 || p[1].Name != "nightly" {
		t.Fatalf("%v %v", p, err)
	}
	if _, err := ParsePipelines("nope"); err == nil {
		t.Fatal("want an error")
	}
}
