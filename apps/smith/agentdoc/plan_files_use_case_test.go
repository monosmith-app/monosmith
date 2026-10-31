package agentdoc

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPlanFilesIsCurrentAfterTheFilesAreWritten(t *testing.T) {
	root := t.TempDir()
	changes, err := PlanFiles(root, sampleFacts(), true)
	if err != nil || len(changes) != 2 {
		t.Fatalf("%v %v", changes, err)
	}
	for _, c := range changes {
		if !c.Differs || c.Exists {
			t.Errorf("%s should be new", c.Path)
		}
		if err := os.WriteFile(filepath.Join(root, c.Path), []byte(c.Content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	again, err := PlanFiles(root, sampleFacts(), true)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range again {
		if c.Differs {
			t.Errorf("%s differs after a write", c.Path)
		}
	}
}

func TestPlanFilesSeesDriftWhenTheWorkspaceChanges(t *testing.T) {
	root := t.TempDir()
	first, _ := PlanFiles(root, sampleFacts(), false)
	_ = os.WriteFile(filepath.Join(root, AgentsFile), []byte(first[0].Content), 0o644)
	f := sampleFacts()
	f.Derived = nil
	again, _ := PlanFiles(root, f, false)
	if len(again) != 1 || !again[0].Differs {
		t.Errorf("got %+v", again)
	}
}

func TestPlanFilesNamesTheFileWithBrokenMarkers(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, AgentsFile), []byte(BeginMarker+"\n"), 0o644)
	_, err := PlanFiles(root, sampleFacts(), false)
	var broken *BrokenFileError
	if !errors.As(err, &broken) || broken.Path != AgentsFile || !errors.Is(err, ErrBrokenMarkers) {
		t.Errorf("err %v", err)
	}
}
