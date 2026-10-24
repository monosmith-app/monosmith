package runner

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestAnInlineRunPrintsTheCommandItsOutcomeAndWaitsForAKey(t *testing.T) {
	var out bytes.Buffer
	paused := false
	clock := time.Unix(0, 0)
	i := NewInlineWith([]string{"run", "build", "-p", "kit"}, func(_ io.Reader, o, _ io.Writer) (int, error) {
		_, _ = o.Write([]byte("compiling\r\n"))
		clock = clock.Add(2 * time.Second)
		return 0, nil
	}, func() { paused = true }).WithClock(func() time.Time { return clock })
	i.SetStdout(&out)
	if err := i.Run(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"$ smith run build -p kit", "compiling", "✔ done in 2s · smith run build -p kit", "Press any key to return"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if !paused || i.Result.Exit.Code != 0 || i.Result.Took != 2*time.Second {
		t.Fatalf("result %+v paused=%v", i.Result, paused)
	}
}

func TestAFailingInlineRunReportsItsExitCode(t *testing.T) {
	var out bytes.Buffer
	i := NewInlineWith([]string{"run", "test"}, func(io.Reader, io.Writer, io.Writer) (int, error) { return 3, nil }, nil)
	i.SetStdout(&out)
	_ = i.Run()
	if !strings.Contains(out.String(), "✖ failed (exit 3)") || i.Result.Exit.Code != 3 {
		t.Fatalf("%q %+v", out.String(), i.Result)
	}
}
