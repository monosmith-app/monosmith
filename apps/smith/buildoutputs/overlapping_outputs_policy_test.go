package buildoutputs

import (
	"testing"
)

func task(id, root string, outputs ...string) TaskOutputs {
	return TaskOutputs{ID: id, Name: lastPart(id), Root: root, Outputs: outputs}
}

func lastPart(id string) string {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == ':' {
			return id[i+1:]
		}
	}
	return id
}

func lines(overlaps []Overlap) []string {
	var out []string
	for _, o := range overlaps {
		out = append(out, o.Line())
	}
	return out
}

func TestOverlapsNamesTasksThatWriteTheSameFolder(t *testing.T) {
	got := lines(Overlaps([]TaskOutputs{
		task("azure-host:package", "apps/azure-host", "../../dist/drop"),
		task("azure-host:build", "apps/azure-host", "../../dist/apps/azure-host"),
		task("studio-ui:package", "apps/studio-ui", "../../dist/drop"),
		task("studio-ui:build", "apps/studio-ui", "../../dist/apps/studio-ui"),
	}, nil))
	if len(got) != 1 || got[0] != "azure-host:package and studio-ui:package both write dist/drop" {
		t.Errorf("got %v", got)
	}
}

func TestOverlapsNamesAFolderThatHoldsAnother(t *testing.T) {
	got := lines(Overlaps([]TaskOutputs{task("a:package", "libs/a", "../../dist/drop"), task("b:package", "libs/b", "../../dist/drop/b")}, nil))
	if len(got) != 1 || got[0] != "a:package and b:package both write dist/drop and dist/drop/b" {
		t.Errorf("got %v", got)
	}
}

func TestOverlapsLeavesSeparateFoldersAlone(t *testing.T) {
	if got := Overlaps([]TaskOutputs{task("a:package", "libs/a", "../../dist/drop/a"), task("b:package", "libs/b", "../../dist/drop/b")}, nil); len(got) != 0 {
		t.Errorf("got %v", got)
	}
	// One task naming a folder twice is not a clash with itself.
	if got := Overlaps([]TaskOutputs{task("c:build", "libs/c", "dist", "dist/types")}, nil); len(got) != 0 {
		t.Errorf("got %v", got)
	}
}

func TestASharedOutputFolderMayBeFilledByManyTasks(t *testing.T) {
	tasks := []TaskOutputs{task("a:package", "apps/a", "../../dist/drop"), task("b:package", "apps/b", "../../dist/drop")}
	if got := Overlaps(tasks, nil); len(got) != 1 {
		t.Fatalf("without the setting it is reported: %v", got)
	}
	if got := Overlaps(tasks, []string{"dist/drop"}); len(got) != 0 {
		t.Errorf("with the setting it is not: %v", got)
	}
}

func TestLongRunningTasksAreNotCompared(t *testing.T) {
	for name, want := range map[string]bool{"watch": true, "dev": true, "start": true, "serve": true, "build:watch": true, "serve-docs": true, "build": false, "build-dev": false, "build:uat": false, "package": false} {
		if got := LongRunning(name); got != want {
			t.Errorf("LongRunning(%q) = %v, want %v", name, got, want)
		}
	}
	if got := Overlaps([]TaskOutputs{task("models:build", "libs/models", "dist"), task("models:watch", "libs/models", "dist")}, nil); len(got) != 0 {
		t.Errorf("models:watch is not reported: %v", got)
	}
}
