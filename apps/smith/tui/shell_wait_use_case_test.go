package tui

import (
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
)

func TestTheClockRunsOnlyWhileTheOpenTabWaitsForALoad(t *testing.T) {
	s := Shell{tab: commands.TabHome, doctor: doctorview.Model{Loading: true}}
	s.trackWait()
	if s.waitFor != "doctor" || s.waitSince.IsZero() || !s.animating() {
		t.Fatalf("a doctor run in flight on Home: %q %v", s.waitFor, s.waitSince)
	}
	first := s.waitSince
	s.tab = commands.TabDoctor
	s.trackWait()
	if !s.waitSince.Equal(first) {
		t.Fatal("Home and Doctor wait for the same run, so the clock keeps going")
	}
	s.waitSince = time.Now().Add(-9 * time.Second)
	if s.waited() < 9*time.Second {
		t.Fatalf("waited %v", s.waited())
	}
	s.doctor = doctorview.Model{}
	s.trackWait()
	if s.waitFor != "" || s.waited() != 0 || s.animating() {
		t.Fatalf("the run ended: %q %v", s.waitFor, s.waited())
	}
}
