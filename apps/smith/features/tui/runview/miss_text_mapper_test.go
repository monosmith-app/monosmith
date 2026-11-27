package runview

import (
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/features/runevents"
)

func TestTheMissSaysWhyInWords(t *testing.T) {
	for _, c := range []struct {
		m    *runevents.Miss
		want string
	}{
		{nil, ""},
		{&runevents.Miss{Reason: runevents.MissNoEntry}, "no entry yet"},
		{&runevents.Miss{Reason: runevents.MissKey}, "the key changed"},
		{&runevents.Miss{Reason: runevents.MissOff}, "the cache was off (--no-cache)"},
		{&runevents.Miss{Reason: runevents.MissInputs, Files: 1}, "1 file changed"},
		{&runevents.Miss{Reason: runevents.MissInputs, Files: 2, Changed: []string{"command", "unknown"}}, "2 files changed, the command changed"},
		{&runevents.Miss{Reason: runevents.MissInputs}, "the inputs changed"},
	} {
		if got := MissText(c.m); got != c.want {
			t.Errorf("%+v: %q, want %q", c.m, got, c.want)
		}
	}
}

func TestAFinishedEventCarriesTheMissToTheTask(t *testing.T) {
	time0 := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	r := NewRun([]string{"run", "build"}, []string{"web:build"}, time0)
	r.Feed(`{"event":"task.started","task":"web:build"}`, time0)
	r.Feed(`{"event":"task.finished","task":"web:build","status":"passed","durationMs":1200,"miss":{"reason":"inputs-changed","files":2}}`, time0)
	if got := r.Task("web:build").Miss; got != "2 files changed" {
		t.Fatalf("%q", got)
	}
}
