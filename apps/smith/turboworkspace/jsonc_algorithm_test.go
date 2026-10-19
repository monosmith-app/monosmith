package turboworkspace

import (
	"encoding/json"
	"testing"
)

func TestStripJSONC(t *testing.T) {
	in := `{
  // a line comment
  "url": "https://x.dev/a//b", /* block */
  "list": ["a", "b",],
  "esc": "quote \" // not a comment",
}`
	var got struct {
		URL  string   `json:"url"`
		List []string `json:"list"`
		Esc  string   `json:"esc"`
	}
	if err := json.Unmarshal(StripJSONC([]byte(in)), &got); err != nil {
		t.Fatal(err)
	}
	if got.URL != "https://x.dev/a//b" || len(got.List) != 2 || got.Esc != `quote " // not a comment` {
		t.Fatalf("%+v", got)
	}
}
