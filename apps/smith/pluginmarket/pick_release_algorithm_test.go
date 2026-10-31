package pluginmarket

import (
	"strings"
	"testing"
)

func TestPickTakesTheNewestStableVersionByNumber(t *testing.T) {
	idx := mustIndex(t, sampleIndexJSON)
	_, rel, err := Pick(idx, "widget", "")
	if err != nil || rel.Version != "1.10.0" {
		t.Fatalf("got %+v, %v (1.10.0 is newer than 1.9 and the beta is skipped)", rel, err)
	}
}

func TestPickTakesAPreReleaseWhenThereIsNothingElse(t *testing.T) {
	text := strings.Replace(sampleIndexJSON, `"1.0.0": {`, `"1.0.0-rc.1": {`, 1)
	text = strings.Replace(text, `"1.10.0"`, `"1.0.0-rc.2"`, 1)
	text = strings.Replace(text, `"2.0.0-beta.1"`, `"0.9.0-x"`, 1)
	_, rel, err := Pick(mustIndex(t, text), "widget", "")
	if err != nil || rel.Version != "1.0.0-rc.2" {
		t.Fatalf("got %+v, %v", rel, err)
	}
}

func TestPickAnExactVersionIgnoresALeadingV(t *testing.T) {
	idx := mustIndex(t, sampleIndexJSON)
	for _, v := range []string{"2.0.0-beta.1", "v1.0.0"} {
		if _, rel, err := Pick(idx, "widget", v); err != nil || rel.Version != strings.TrimPrefix(v, "v") {
			t.Fatalf("%s: got %+v, %v", v, rel, err)
		}
	}
}

func TestPickSaysWhatExists(t *testing.T) {
	idx := mustIndex(t, sampleIndexJSON)
	if _, _, err := Pick(idx, "gadget", ""); err == nil || !strings.Contains(err.Error(), `no plugin "gadget"`) {
		t.Fatalf("got %v", err)
	}
	if _, _, err := Pick(idx, "widget", "9.9.9"); err == nil || !strings.Contains(err.Error(), "1.10.0, 1.0.0") {
		t.Fatalf("got %v", err)
	}
}

func TestVersionsAreNewestFirst(t *testing.T) {
	got := strings.Join(Versions(mustIndex(t, sampleIndexJSON).Plugins["widget"]), " ")
	if got != "2.0.0-beta.1 1.10.0 1.0.0" {
		t.Fatal(got)
	}
}
