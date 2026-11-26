package tui

import "testing"

func TestAClickOnAFilterButtonOfDoctorHidesAndShowsItsChecks(t *testing.T) {
	l := newLab(t, true, 170, 30)
	l.keys("3")
	l.has("× 2 failed (x)", "go.work lists every module", "x hide failed")
	l.clickText("× 2 failed (x)")
	l.hasNot("go.work lists every module")
	l.has("No checks match the filters. Turn one on (x, w or p).", "x show failed")
	l.clickText("× 2 failed (x)")
	l.has("go.work lists every module", "x hide failed")
}

func TestTheFilterKeysDoWhatTheButtonsDo(t *testing.T) {
	l := newLab(t, true, 170, 30)
	l.keys("3", "x")
	l.hasNot("go.work lists every module")
	l.keys("x", "p")
	l.has("go.work lists every module", "git is installed")
}
