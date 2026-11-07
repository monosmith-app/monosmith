package tui

import "testing"

func TestAClickOnAFilterButtonOfDoctorHidesAndShowsItsChecks(t *testing.T) {
	l := newLab(t, true, 170, 30)
	l.keys("4")
	l.has("× 2 failed (x)", "go.work lists every module", "hide failed (x)")
	l.clickText("× 2 failed (x)")
	l.hasNot("go.work lists every module")
	l.has("No checks match the filters. Turn one on (x, w or p).", "show failed (x)")
	l.clickText("× 2 failed (x)")
	l.has("go.work lists every module", "hide failed (x)")
}

func TestTheFilterKeysDoWhatTheButtonsDo(t *testing.T) {
	l := newLab(t, true, 170, 30)
	l.keys("4", "x")
	l.hasNot("go.work lists every module")
	l.keys("x", "p")
	l.has("go.work lists every module", "git is installed")
}
