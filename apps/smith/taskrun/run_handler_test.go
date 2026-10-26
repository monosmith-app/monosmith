package taskrun

import (
	"bytes"
	"testing"
)

func TestJSONSummaryKeepsStdoutFreeOfTaskOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	_, _ = textOutput(&stdout, &stderr, true).Write([]byte("log line\n"))
	if stdout.Len() != 0 || stderr.String() != "log line\n" {
		t.Fatalf("stdout %q stderr %q", stdout.String(), stderr.String())
	}
}

func TestTextRunsStillPrintTaskOutputOnStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	_, _ = textOutput(&stdout, &stderr, false).Write([]byte("log line\n"))
	if stdout.String() != "log line\n" || stderr.Len() != 0 {
		t.Fatalf("stdout %q stderr %q", stdout.String(), stderr.String())
	}
}
