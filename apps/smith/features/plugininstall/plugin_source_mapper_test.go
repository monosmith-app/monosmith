package plugininstall

import "testing"

func TestParseSource(t *testing.T) {
	cases := map[string]Source{
		"./widget.smith-plugin.zip":                      {Kind: KindLocal, Location: "./widget.smith-plugin.zip"},
		`C:\plugins\widget`:                              {Kind: KindLocal, Location: `C:\plugins\widget`},
		"https://example.com/w.smith-plugin.zip":         {Kind: KindURL, Location: "https://example.com/w.smith-plugin.zip"},
		"http://localhost:8080/w.smith-plugin.zip":       {Kind: KindURL, Location: "http://localhost:8080/w.smith-plugin.zip"},
		"github:acme/widget@v1.2.3":                      {Kind: KindGitHub, Owner: "acme", Repo: "widget", Tag: "v1.2.3"},
		"github:acme/smith.plugin@v1#w.smith-plugin.zip": {Kind: KindGitHub, Owner: "acme", Repo: "smith.plugin", Tag: "v1", Asset: "w.smith-plugin.zip"},
	}
	for arg, want := range cases {
		got, err := ParseSource(arg)
		if err != nil || got != want {
			t.Errorf("%s: got %+v, %v", arg, got, err)
		}
	}
	for _, bad := range []string{"", "  ", "http://example.com/w.zip", "ftp://x/y", "github:acme/widget", "github:acme@v1", "github:a/b/c@v1", "github:acme/widget@", "github:acme/widget@v1#../x"} {
		if _, err := ParseSource(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestGitHubSourceString(t *testing.T) {
	s, _ := ParseSource("github:acme/widget@v1#a.zip")
	if s.String() != "github:acme/widget@v1#a.zip" {
		t.Fatal(s.String())
	}
}
