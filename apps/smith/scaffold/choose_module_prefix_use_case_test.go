package scaffold

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestChooseModulePrefix(t *testing.T) {
	cases := []struct {
		name  string
		req   PrefixRequest
		input string
		want  string
		said  string
	}{
		{"the flag wins and says nothing", PrefixRequest{Given: "github.com/acme/x"}, "", "github.com/acme/x", ""},
		{"not Go: nothing to decide", PrefixRequest{Language: "node"}, "", "", ""},
		{"no terminal: placeholder, announced with the flag", PrefixRequest{}, "", "example.com/acme", "--module-prefix"},
		{"terminal: typed prefix", PrefixRequest{Interactive: true}, "type\ngithub.com/me/acme\n", "github.com/me/acme", "Go module prefix: github.com/me/acme"},
		{"terminal: the placeholder", PrefixRequest{Interactive: true}, "example\n", "example.com/acme", "Go module prefix: example.com/acme"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			c.req.Name, c.req.Dir = "acme", t.TempDir()
			c.req.In, c.req.Out = bufio.NewReader(strings.NewReader(c.input)), &out
			if got := ChooseModulePrefix(c.req); got != c.want {
				t.Fatalf("got %q, want %q\n%s", got, c.want, out.String())
			}
			if !strings.Contains(out.String(), c.said) {
				t.Fatalf("output lacks %q:\n%s", c.said, out.String())
			}
		})
	}
}
