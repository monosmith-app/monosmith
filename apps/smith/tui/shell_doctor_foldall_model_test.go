package tui

import (
	"fmt"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/doctorview"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func foldLab(t *testing.T, w, h int) *lab {
	t.Helper()
	l := newLab(t, true, w, h)
	var fs []doctor.Finding
	for i := range 6 {
		fs = append(fs, doctor.Finding{Finding: protocol.Finding{Check: fmt.Sprintf("check %d", i), Problem: fmt.Sprintf("problem %d", i), Project: fmt.Sprintf("pkg%d", i), Warning: true, Detail: "d", Remedy: "r"}, Source: "node"})
	}
	l.send(doctorview.Result{Report: doctor.Report{Warnings: 6, Findings: fs}})
	l.keys("3")
	return l
}

func TestTheFooterNamesWhatTheFoldAllKeyDoesNext(t *testing.T) {
	for _, size := range [][2]int{{140, 34}, {110, 30}, {100, 30}} {
		l := foldLab(t, size[0], size[1])
		l.has("collapse all (E)")
		l.keys("E")
		l.has("expand all (E)")
		l.hasNot("collapse all (E)")
		l.keys("E")
		l.has("collapse all (E)")
	}
}
