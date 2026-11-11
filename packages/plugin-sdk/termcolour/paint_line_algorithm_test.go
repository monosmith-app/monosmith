package termcolour

import "testing"

func TestPaintColoursTheResultLinesOfGoTestAndMSBuild(t *testing.T) {
	cases := []struct{ line, code string }{
		{"--- FAIL: TestX (0.00s)", Red},
		{"    --- FAIL: TestX/sub (0.00s)", Red},
		{"FAIL\tgithub.com/x/y\t0.012s", Red},
		{"FAIL", Red},
		{"panic: runtime error: index out of range", Red},
		{"error: could not compile `app`", Red},
		{"error[E0425]: cannot find value `x`", Red},
		{"src/a.ts(1,7): error TS2322: Type 'string' is not assignable", Red},
		{"Program.cs(3,1): error CS0103: The name 'x' does not exist", Red},
		{"warning: unused variable: `x`", Yellow},
		{"Program.cs(3,1): warning CS0168: The variable 'e' is declared but never used", Yellow},
		{"--- SKIP: TestY (0.00s)", Dim},
		{"?   \tgithub.com/x/z\t[no test files]", Dim},
		{"--- PASS: TestX (0.00s)", Green},
		{"ok  \tgithub.com/x/y\t0.012s", Green},
		{"PASS", Green},
		{"Build FAILED.", Red},
		{"Build succeeded.", Green},
		// The Rust test harness, as cargo test prints it into a pipe.
		{"test tests::fails ... FAILED", Red},
		{"test result: FAILED. 1 passed; 1 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s", Red},
		{"thread 'tests::fails' (17300) panicked at src\\lib.rs:15:9:", Red},
		{"test tests::later ... ignored", Dim},
		{"test tests::passes ... ok", Green},
		{"test result: ok. 2 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.00s", Green},
	}
	for _, c := range cases {
		if got, want := Paint(c.line), Wrap(c.line, c.code); got != want {
			t.Errorf("%q: got %q, want %q", c.line, got, want)
		}
	}
}

func TestPaintLeavesOtherLinesAndColouredLinesAlone(t *testing.T) {
	for _, line := range []string{
		"", "=== RUN   TestX", "    x_test.go:12: got 1, want 2", "okay then", "PASSED",
		"\x1b[31m--- FAIL: TestX\x1b[0m", // the tool coloured it already
	} {
		if got := Paint(line); got != line {
			t.Errorf("%q changed to %q", line, got)
		}
	}
}

func TestPaintNeverChangesTheCharactersOfALine(t *testing.T) {
	for _, line := range []string{"--- FAIL: TestX (0.00s)", "ok  \tx\t1s", "warning: x", "plain"} {
		if got := Strip(Paint(line)); got != line {
			t.Errorf("%q became %q without its colour", line, got)
		}
	}
}

func TestLineEndsAColouredLineWithAReset(t *testing.T) {
	if got := Line("\x1b[33mstill yellow", true); got != "\x1b[33mstill yellow\x1b[0m" {
		t.Errorf("got %q", got)
	}
	if got := Line("\x1b[33my\x1b[m", true); got != "\x1b[33my\x1b[m" {
		t.Errorf("a line that resets itself keeps its bytes: %q", got)
	}
	if got := Line("plain", true); got != "plain" {
		t.Errorf("a plain line gets no reset: %q", got)
	}
}

func TestLineIsPlainWhenColourIsOff(t *testing.T) {
	if got := Line("\x1b[31m--- FAIL: TestX\x1b[0m", false); got != "--- FAIL: TestX" {
		t.Errorf("got %q", got)
	}
	if got := Line("--- FAIL: TestX", false); got != "--- FAIL: TestX" {
		t.Errorf("smith paint must not apply to a plain destination: %q", got)
	}
	if got := Line("--- FAIL: TestX", true); got != Wrap("--- FAIL: TestX", Red) {
		t.Errorf("got %q", got)
	}
}
