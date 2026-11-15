package languageserver

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteFrame(&buf, []byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	if err := WriteFrame(&buf, []byte(`{"b":"é"}`)); err != nil {
		t.Fatal(err)
	}
	r := bufio.NewReader(&buf)
	for _, want := range []string{`{"a":1}`, `{"b":"é"}`} {
		got, err := ReadFrame(r)
		if err != nil || string(got) != want {
			t.Fatalf("got %q, %v; want %q", got, err, want)
		}
	}
	if _, err := ReadFrame(r); !errors.Is(err, io.EOF) {
		t.Fatalf("want EOF, got %v", err)
	}
}

func TestReadFrameIgnoresOtherHeadersAndCase(t *testing.T) {
	in := "Content-Type: application/vscode-jsonrpc\r\ncontent-length: 2\r\n\r\n{}"
	got, err := ReadFrame(bufio.NewReader(strings.NewReader(in)))
	if err != nil || string(got) != "{}" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestReadFrameRejectsMalformed(t *testing.T) {
	for name, in := range map[string]string{
		"no length":  "X: 1\r\n\r\n{}",
		"bad length": "Content-Length: abc\r\n\r\n",
		"short body": "Content-Length: 10\r\n\r\n{}",
	} {
		if _, err := ReadFrame(bufio.NewReader(strings.NewReader(in))); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
