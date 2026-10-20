package doctorview

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func finding(check, detail string) doctor.Finding {
	return doctor.Finding{Finding: protocol.Finding{Check: check, Detail: detail}}
}

func TestGroupByProjectFilesAFindingUnderTheProjectItNames(t *testing.T) {
	projects := []ProjectRef{{Name: "web", Root: "apps/web"}, {Name: "api", Root: "apps/api"}, {Name: "go", Root: "libs/go"}}
	findings := []doctor.Finding{
		finding("Go is installed", "go is not on PATH"),
		finding("Docker registries of api agree", ""),
		finding("go.work lists apps/web", "apps/web is missing"),
		finding("the clone has full history", ""),
		finding("Docker Hub short description of web fits", ""),
	}
	got := GroupByProject(findings, projects)
	want := []Group{
		{Indexes: []int{0, 3}},
		{Project: "web", Indexes: []int{2, 4}},
		{Project: "api", Indexes: []int{1}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("groups\n got %+v\nwant %+v", got, want)
	}
}

func TestGroupByProjectDoesNotMistakeAFileNameForAProject(t *testing.T) {
	// "go.work" is one token, so it is not the project called go.
	projects := []ProjectRef{{Name: "go", Root: "libs/go"}}
	got := GroupByProject([]doctor.Finding{finding("go.work exists", "the workspace has Go projects but no go.work")}, projects)
	if len(got) != 1 || got[0].Project != "" {
		t.Fatalf("a workspace finding was filed under a project: %+v", got)
	}
}

func TestGroupByProjectLeavesOnlyTheWorkspaceWhenNoProjectIsNamed(t *testing.T) {
	got := GroupByProject([]doctor.Finding{finding("git is installed", "")}, nil)
	if len(got) != 1 || got[0].Project != "" || len(got[0].Indexes) != 1 {
		t.Fatalf("%+v", got)
	}
	if got := GroupByProject(nil, nil); len(got) != 0 {
		t.Fatalf("no findings, no groups: %+v", got)
	}
}
