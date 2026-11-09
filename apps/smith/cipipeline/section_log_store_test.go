package cipipeline

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// clockStep is a clock that moves one second a call, so a section lasts 1s.
func clockStep() func() time.Time {
	t := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return func() time.Time { t = t.Add(time.Second); return t }
}

func testSections(out *bytes.Buffer, icons theme.Icons, colour bool) *SectionLog {
	return &SectionLog{out: out, icons: icons, colour: colour, now: clockStep()}
}

func noCI(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("TF_BUILD", "")
}

func TestSectionsAreInOrderAndNeverMix(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, false)
	for _, name := range []string{"node", "python", "go"} {
		l.Begin(name, 2)
		fmt.Fprintf(&out, "%s output 1\n%s output 2\n", name, name)
		l.End(nil)
	}
	got := out.String()
	last := -1
	for _, name := range []string{"══ node (2 projects) ═", "node output 2", "✓ node done in 1.0s", "══ python (2 projects)", "python output 2", "✓ python done in 1.0s", "══ go (2 projects)", "go output 2", "✓ go done in 1.0s"} {
		i := strings.Index(got, name)
		if i <= last {
			t.Fatalf("%q is missing or out of order in:\n%s", name, got)
		}
		last = i
	}
}

func TestRuleFitsTheBannerWidth(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, false)
	l.Begin("node", 1)
	line := strings.SplitN(out.String(), "\n", 2)[0]
	if n := len([]rune(line)); n > 80 || n != bannerWidth {
		t.Fatalf("rule is %d columns: %q", n, line)
	}
}

func TestASCIIFallbackHasNoUnicode(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.ASCII, false)
	l.Begin("node", 1)
	l.End(nil)
	l.Begin("go", 2)
	l.End(errors.New("boom"))
	l.Summary()
	got := out.String()
	for _, want := range []string{"== node (1 project) ==", "[v] node done in 1.0s", "[x] go failed: boom", "[v] done", "[x] failed"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	for _, r := range strings.ReplaceAll(got, "→", "") {
		if r > 127 {
			t.Fatalf("non-ASCII rune %q in:\n%s", r, got)
		}
	}
}

func TestNoColourWhenOff(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, false)
	l.Begin("node", 1)
	l.End(errors.New("boom"))
	l.Summary()
	if strings.Contains(out.String(), "\x1b[") {
		t.Fatalf("colour escape in plain output:\n%q", out.String())
	}
}

func TestColourWhenOn(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, true)
	l.Begin("node", 1)
	l.End(nil)
	if !strings.Contains(out.String(), "\x1b[32m✓\x1b[0m node done") {
		t.Fatalf("the done glyph is not green:\n%q", out.String())
	}
}

func TestFailureSectionNamesCauseAndRemedy(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, false)
	l.Begin("python", 3)
	l.End(errors.New("uv: no matching distribution\nsecond line"))
	got := out.String()
	if !strings.Contains(got, "× python failed: uv: no matching distribution\n→ ") || strings.Contains(got, "second line") {
		t.Fatalf("failure lines are wrong:\n%s", got)
	}
}

func TestSummaryHasOneRowPerToolchainAndCountsFailures(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, false)
	l.Begin("node", 2)
	l.End(nil)
	l.Begin("go", 1)
	l.End(errors.New("x"))
	if failed := l.Summary(); failed != 1 {
		t.Fatalf("failed = %d, want 1", failed)
	}
	got := out.String()
	for _, want := range []string{"Toolchain  Projects  Status", "node              2  ✓ done", "go                1  × failed"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestSummaryIsEmptyWithoutSections(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	if testSections(&out, theme.Unicode, false).Summary() != 0 || out.Len() != 0 {
		t.Fatalf("printed %q", out.String())
	}
}

func TestOpenSectionIsClosedAsFailedByTheNextBegin(t *testing.T) {
	noCI(t)
	var out bytes.Buffer
	l := testSections(&out, theme.Unicode, false)
	l.Begin("node", 1)
	l.Begin("go", 1)
	l.End(nil)
	if got := out.String(); !strings.Contains(got, "× node failed") || !strings.Contains(got, "✓ go done") {
		t.Fatalf("unexpected:\n%s", got)
	}
}

func TestCIGroupsReplaceTheRule(t *testing.T) {
	for _, tc := range []struct{ env, open, shut string }{
		{"GITHUB_ACTIONS", "::group::node (1 project)\n", "::endgroup::\n✓ node done"},
		{"TF_BUILD", "##[group]node (1 project)\n", "##[endgroup]\n✓ node done"},
	} {
		t.Run(tc.env, func(t *testing.T) {
			noCI(t)
			t.Setenv(tc.env, "true")
			var out bytes.Buffer
			l := testSections(&out, theme.Unicode, false)
			l.Begin("node", 1)
			fmt.Fprintln(&out, "npm output")
			l.End(nil)
			got := out.String()
			if !strings.HasPrefix(got, tc.open) || !strings.Contains(got, tc.shut) || strings.Contains(got, "══") {
				t.Fatalf("CI section is wrong:\n%s", got)
			}
		})
	}
}
