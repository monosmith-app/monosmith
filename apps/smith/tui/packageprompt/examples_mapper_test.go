package packageprompt

import "testing"

func TestTheExamplesFollowTheProjectsLanguage(t *testing.T) {
	p := NewInstall("api", []string{"api", "web"}).WithLanguages(map[string]string{"api": "python", "web": "node"})
	if got := p.input.Placeholder; got != ExampleNames("python") {
		t.Fatalf("the first project's examples are %q", got)
	}
	if got := p.step(1).input.Placeholder; got != ExampleNames("node") {
		t.Fatalf("after the project changed, the examples are %q", got)
	}
}
