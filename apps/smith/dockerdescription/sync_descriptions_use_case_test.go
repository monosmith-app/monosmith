package dockerdescription

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/libs/plugin-docker/publishing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func project(name string) *projectgraph.Project {
	return &projectgraph.Project{Project: protocol.Project{Name: name}}
}

// fakeSync answers per project name and records what it was asked.
func fakeSync(t *testing.T, answers map[string]publishing.DescriptionResult) *[]string {
	t.Helper()
	var asked []string
	real := syncDescription
	syncDescription = func(_ string, p protocol.Project, dryRun bool) (publishing.DescriptionResult, error) {
		asked = append(asked, p.Name)
		if p.Name == "broken" {
			return publishing.DescriptionResult{}, errors.New("registries contradict each other")
		}
		return answers[p.Name], nil
	}
	t.Cleanup(func() { syncDescription = real })
	return &asked
}

func TestEachAppWithADescriptionIsSetAndOthersStaySilent(t *testing.T) {
	asked := fakeSync(t, map[string]publishing.DescriptionResult{
		"api": {Set: []string{"docker.io/acme/api"}},
		"web": {Skipped: "no description_from"},
	})
	var out bytes.Buffer
	if err := Run("/ws", []*projectgraph.Project{project("api"), project("web")}, Request{}, &out); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "api: set the Docker Hub description of docker.io/acme/api\n" {
		t.Errorf("output %q", got)
	}
	if strings.Join(*asked, ",") != "api,web" {
		t.Errorf("asked %v", *asked)
	}
}

func TestADryRunSaysWouldAndAProjectCanBeAskedForByName(t *testing.T) {
	fakeSync(t, map[string]publishing.DescriptionResult{
		"api": {Set: []string{"docker.io/acme/api"}, Short: "The API.", Full: "README.md"},
		"web": {Skipped: "no docker.io registry"},
	})
	var out bytes.Buffer
	err := Run("/ws", []*projectgraph.Project{project("api"), project("web")}, Request{Projects: []string{"web"}, DryRun: true}, &out)
	if err != nil || out.String() != "web: nothing to set (no docker.io registry)\n" {
		t.Errorf("%v %q", err, out.String())
	}
	out.Reset()
	if err := Run("/ws", []*projectgraph.Project{project("api")}, Request{DryRun: true}, &out); err != nil || !strings.Contains(out.String(), "would set the Docker Hub description of docker.io/acme/api") {
		t.Errorf("%v %q", err, out.String())
	}
	if got := out.String(); !strings.Contains(got, "short description: The API.") || !strings.Contains(got, "full description: README.md") {
		t.Errorf("a dry run shows both descriptions: %q", got)
	}
}

func TestAFailureIsAWarningUnlessStrict(t *testing.T) {
	fakeSync(t, map[string]publishing.DescriptionResult{"api": {Failed: []string{"docker.io/acme/api: 403 Forbidden"}}})
	projects := []*projectgraph.Project{project("api")}
	var out bytes.Buffer
	if err := Run("/ws", projects, Request{}, &out); err != nil || !strings.Contains(out.String(), "could not set the Docker Hub description: docker.io/acme/api: 403 Forbidden") {
		t.Errorf("not strict: %v %q", err, out.String())
	}
	if err := Run("/ws", projects, Request{Strict: true}, &out); !errors.Is(err, ErrFailed) {
		t.Errorf("strict: %v", err)
	}
}

func TestContradictingSettingsAreAnErrorNamingTheApp(t *testing.T) {
	fakeSync(t, nil)
	err := Run("/ws", []*projectgraph.Project{project("broken")}, Request{}, &bytes.Buffer{})
	if err == nil || !strings.HasPrefix(err.Error(), "broken: ") {
		t.Errorf("%v", err)
	}
}

func TestSayingSoWhenNoAppNamesADescription(t *testing.T) {
	fakeSync(t, map[string]publishing.DescriptionResult{"web": {Skipped: "no description_from"}})
	var out bytes.Buffer
	if err := Run("/ws", []*projectgraph.Project{project("web")}, Request{}, &out); err != nil || !strings.Contains(out.String(), "no app has a description_from file or a short description") {
		t.Errorf("%v %q", err, out.String())
	}
}
