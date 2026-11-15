package languageserver

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ReadFrame reads one LSP message body: header lines up to a blank line, then
// Content-Length bytes. It returns io.EOF when the stream ends between
// messages, and an error when a header or the body is malformed.
func ReadFrame(r *bufio.Reader) ([]byte, error) {
	length := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) && line == "" {
				return nil, io.EOF
			}
			return nil, fmt.Errorf("lsp: reading header: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			n, convErr := strconv.Atoi(strings.TrimSpace(value))
			if convErr != nil || n < 0 {
				return nil, fmt.Errorf("lsp: bad Content-Length %q", value)
			}
			length = n
		}
	}
	if length < 0 {
		return nil, errors.New("lsp: message without Content-Length")
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("lsp: reading body: %w", err)
	}
	return body, nil
}

// WriteFrame writes body as one LSP message with its Content-Length header.
func WriteFrame(w io.Writer, body []byte) error {
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err := w.Write(body)
	return err
}
