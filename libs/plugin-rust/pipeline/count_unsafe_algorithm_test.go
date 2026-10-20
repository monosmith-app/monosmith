package pipeline

import "testing"

func TestCountUnsafeCountsBlocksFunctionsAndImpls(t *testing.T) {
	src := "unsafe fn a() {}\nfn b() { unsafe { x() } }\nunsafe impl Send for T {}\nunsafe trait U {}\n"
	if got := CountUnsafe(src); got != 4 {
		t.Errorf("got %d", got)
	}
}

func TestCountUnsafeIgnoresCommentsAndStrings(t *testing.T) {
	src := "// unsafe { }\n/* unsafe { /* nested unsafe { */ unsafe { } */\nlet s = \"unsafe { \\\" unsafe {\";\nlet q = '\"';\nfn c() { unsafe { } }\n"
	if got := CountUnsafe(src); got != 1 {
		t.Errorf("got %d", got)
	}
}
