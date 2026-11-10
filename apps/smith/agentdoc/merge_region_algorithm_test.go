package agentdoc

import (
	"errors"
	"strings"
	"testing"
)

const region = BeginMarker + "\nnew\n" + EndMarker + "\n"

func TestMergeMakesANewFileFromTheHeader(t *testing.T) {
	got, err := Merge("", false, "# H\n", region)
	if err != nil || got != "# H\n\n"+region {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestMergeKeepsHandWrittenTextOutsideTheMarkers(t *testing.T) {
	old := "# Mine\n\nbefore\n\n" + BeginMarker + "\nold\n" + EndMarker + "\n\nafter\n"
	got, err := Merge(old, true, "", region)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Mine\n\nbefore\n\n" + region + "\nafter\n"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	again, _ := Merge(got, true, "", region)
	if again != got {
		t.Error("a second merge changed the file")
	}
}

func TestMergeAppendsToAFileWithNoMarkers(t *testing.T) {
	got, err := Merge("# Mine\n", true, "", region)
	if err != nil || got != "# Mine\n\n"+region {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestMergeKeepsCRLF(t *testing.T) {
	old := "# Mine\r\n" + BeginMarker + "\r\nold\r\n" + EndMarker + "\r\n"
	got, err := Merge(old, true, "", region)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") {
		t.Errorf("a bare LF in %q", got)
	}
}

func TestMergeRefusesBrokenMarkers(t *testing.T) {
	for _, old := range []string{BeginMarker + "\n", EndMarker + "\n" + BeginMarker + "\n"} {
		if _, err := Merge(old, true, "", region); !errors.Is(err, ErrBrokenMarkers) {
			t.Errorf("%q: err %v", old, err)
		}
	}
}
