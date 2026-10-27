package cipipeline

import (
	"fmt"
	"strings"
	"testing"
)

func TestATableAlignsColumnsWithSpacesAndRightAlignsNumbers(t *testing.T) {
	table := resultTable{
		Headers: []string{"Project", "Result", "Time"}, Right: []bool{false, false, true},
		Rows: []tableRow{{Cells: []string{"core", "✓ passed", "12s"}}, {Cells: []string{"atlassian", "✗ failed", "3m 12s"}, Bad: true}},
	}
	want := "Project    Result      Time\ncore       ✓ passed     12s\natlassian  ✗ failed  3m 12s"
	if got := strings.Join(table.Text("passed"), "\n"); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if md := table.Markdown(); md[1] != "| --- | --- | ---: |" || md[3] != "| atlassian | ✗ failed | 3m 12s |" {
		t.Errorf("markdown: %v", md)
	}
}

func TestALongTableKeepsTheBadRowsAndCountsTheRest(t *testing.T) {
	table := resultTable{Headers: []string{"Project"}}
	for i := 0; i < 40; i++ {
		table.Rows = append(table.Rows, tableRow{Cells: []string{fmt.Sprintf("p%02d", i)}, Bad: i == 39})
	}
	lines := table.Text("passed")
	if len(lines) != maxTableRows+2 || lines[len(lines)-1] != "... 10 more passed" || !strings.Contains(strings.Join(lines, "\n"), "p39") {
		t.Errorf("%d lines, last %q", len(lines), lines[len(lines)-1])
	}
}
