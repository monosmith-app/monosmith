package runhistory

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
)

func TestRerunFailedBuildsAGridOfTasksAndProjects(t *testing.T) {
	got := RerunFailed([]string{"kit:test", "web:lint", "kit:lint", "kit:test", "broken"})
	want := []string{"run", "test", "lint", flagvocabulary.Arg("project"), "kit,web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if RerunFailed(nil) != nil || RerunFailed([]string{"broken"}) != nil {
		t.Fatal("no task, no command")
	}
}
