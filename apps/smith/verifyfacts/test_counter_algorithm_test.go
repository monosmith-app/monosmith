package verifyfacts

import (
	"testing"
)

func countTests(lines ...string) (passed, failed int, unit string, known bool) {
	var c TestCounter
	for _, l := range lines {
		c.Add(l)
	}
	return c.Result()
}

func TestTestCounterReadsTheSummaryOfEachRunner(t *testing.T) {
	cases := []struct {
		name           string
		lines          []string
		passed, failed int
		unit           string
	}{
		{"jest", []string{"Test Suites: 1 failed, 2 passed, 3 total", "Tests:       1 failed, 4 passed, 5 total"}, 4, 1, "tests"},
		{"vitest", []string{" Test Files  1 failed | 2 passed (3)", "      Tests  1 failed | 4 passed (5)"}, 4, 1, "tests"},
		{"pytest", []string{"=== 1 failed, 4 passed in 0.12s ==="}, 4, 1, "tests"},
		{"pytest all green", []string{"4 passed in 0.03s"}, 4, 0, "tests"},
		{"dotnet", []string{"Failed!  - Failed:     2, Passed:    10, Skipped:     0, Total:    12"}, 10, 2, "tests"},
		{"cargo adds the binaries", []string{"test result: ok. 3 passed; 0 failed; 0 ignored", "test result: FAILED. 2 passed; 1 failed; 0 ignored"}, 5, 1, "tests"},
		{"go packages", []string{"ok  \texample.com/a\t0.3s", "ok  \texample.com/b\t0.1s", "?   \texample.com/c\t[no test files]"}, 2, 0, "packages"},
		{"go with a failed test", []string{"ok  \texample.com/a\t0.3s", "--- FAIL: TestX (0.00s)", "FAIL\texample.com/b\t0.2s"}, 1, 1, "packages"},
	}
	for _, c := range cases {
		passed, failed, unit, known := countTests(c.lines...)
		if !known || passed != c.passed || failed != c.failed || unit != c.unit {
			t.Errorf("%s: got %d passed, %d failed, %q, known=%v; want %d, %d, %q", c.name, passed, failed, unit, known, c.passed, c.failed, c.unit)
		}
	}
	if _, _, _, known := countTests("compiling...", "done"); known {
		t.Error("a runner that said nothing readable has counts")
	}
}
