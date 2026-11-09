package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// slowDoctor makes every doctor run of the lab wait for release, so the screen
// can be read while doctor is still checking. The tests send the result
// themselves.
func (l *lab) slowDoctor(first doctorview.Result) (release chan struct{}) {
	release = make(chan struct{})
	sh := l.sh()
	loads := 0
	sh.sources.Doctor = func() doctorview.Result {
		loads++
		if loads == 1 {
			return first
		}
		<-release
		return first
	}
	l.m = sh
	return release
}

func twoFindings() doctorview.Result {
	return doctorview.Result{Report: doctor.Report{Failed: 1, Warnings: 1, Passed: 3, Findings: []doctor.Finding{
		{Finding: protocol.Finding{Check: "first check", Detail: "first is wrong"}},
		{Finding: protocol.Finding{Check: "second check", Warning: true, Detail: "second is odd"}},
	}}}
}

func TestCheckAgainKeepsTheListAndSaysItIsCheckingThenSaysWhatItFound(t *testing.T) {
	l := newLab(t, true, 140, 34)
	release := l.slowDoctor(twoFindings())
	defer close(release)
	l.keys("3", "r")
	l.has("Check again (r)", "first check")
	l.keys("down")
	l.keys("w")
	before := l.sh().doctor.Cursor
	l.keys("r")
	l.has("Checking again…", "first check")
	for _, line := range l.lines() {
		if strings.Contains(line, "Checking again…") && strings.Contains(line, "[Check again") {
			t.Fatalf("the button stays beside the hint: %q", line)
		}
	}
	l.send(twoFindings())
	l.has("Check again (r)", "Checked again: 1 failed, 1 warning")
	l.hasNot("Checking again…")
	if got := l.sh().doctor.Cursor; got != before || !l.sh().doctor.HideWarnings {
		t.Fatalf("cursor %d (was %d), warnings hidden %v: the list lost its place", got, before, l.sh().doctor.HideWarnings)
	}
}

func TestTheCheckAgainButtonAnswersAClick(t *testing.T) {
	l := newLab(t, true, 140, 34)
	release := l.slowDoctor(twoFindings())
	defer close(release)
	l.keys("3", "r")
	l.clickText("Check again (r)")
	if !l.sh().doctor.Loading {
		t.Fatal("the click did not start a check")
	}
}

func TestTheListSaysItIsCheckingAgainRightAfterAFixEnds(t *testing.T) {
	l := newLab(t, true, 140, 34)
	release := l.slowDoctor(doctorview.Result{Report: doctor.Report{Failed: 1, Findings: []doctor.Finding{{Finding: protocol.Finding{
		Check: "a file is missing", Detail: "x is missing",
		Fix: &protocol.Fix{Description: "write x", Safe: true, Files: []protocol.FileEdit{{Path: "x.txt", Content: "x\n"}}},
	}}}}})
	defer close(release)
	l.keys("3", "r", "f", "enter")
	l.has("Checking again…", "a file is missing")
	if !l.sh().doctor.Loading {
		t.Fatal("doctor is out and the list does not say so")
	}
}

func TestTheDoctorFilterButtonsStayAtTheTopWhileTheChecksScroll(t *testing.T) {
	for _, width := range []int{140, 76} {
		l := newLab(t, true, width, 20)
		var res doctorview.Result
		for i := 0; i < 60; i++ {
			res.Report.Findings = append(res.Report.Findings, doctor.Finding{Finding: protocol.Finding{Check: "check number " + strings.Repeat("x", i%5) + string(rune('a'+i%26)), Detail: "wrong"}})
		}
		res.Report.Failed = 60
		sh := l.sh()
		sh.sources.Doctor = func() doctorview.Result { return res }
		l.m = sh
		l.keys("3", "r")
		row := func() int {
			for i, line := range l.lines() {
				if strings.Contains(line, "60 failed") {
					return i
				}
			}
			return -1
		}
		first := row()
		if first < 0 {
			t.Fatalf("width %d: no filter buttons: %s", width, l.view())
		}
		top := l.view()
		for i := 0; i < 30; i++ {
			l.keys("down")
		}
		if l.view() == top {
			t.Fatalf("width %d: the list did not scroll: %s", width, top)
		}
		if got := row(); got != first {
			t.Fatalf("width %d: the buttons moved from row %d to %d (-1 is gone) when the list scrolled", width, first, got)
		}
		l.send(wheel(5, first+4, -1))
		if got := row(); got != first {
			t.Fatalf("width %d: the wheel moved the buttons to row %d", width, got)
		}
	}
}
