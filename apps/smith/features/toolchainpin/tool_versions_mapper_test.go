package toolchainpin

import "testing"

func TestMergeKeepsOtherLinesAndUpdatesPinnedTools(t *testing.T) {
	got := MergeToolVersions("# tools\nnode 18.0.0\nterraform 1.9.0\n", map[string]string{"node": "22", "go": "1.23", "rustc": "1.80"})
	want := "# tools\nnode 22\nterraform 1.9.0\ngo 1.23\nrust 1.80\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	if got := MergeToolVersions("", map[string]string{"node": "22"}); got != "node 22\n" {
		t.Errorf("a new file: %q", got)
	}
}
