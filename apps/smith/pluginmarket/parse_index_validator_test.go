package pluginmarket

import (
	"strings"
	"testing"
)

const digestA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

// sampleIndexJSON lists two plugins; widget has a stable, a newer stable and a pre-release.
const sampleIndexJSON = `{
  "schema": 1,
  "plugins": {
    "widget": {
      "description": "Scaffolds widget projects",
      "roles": ["scaffold"], "kinds": ["widget"], "languages": ["go"],
      "license": "MIT",
      "versions": {
        "1.0.0": {"url": "https://example.com/widget-1.0.0.smith-plugin.zip", "sha256": "` + digestA + `"},
        "1.10.0": {"url": "https://example.com/widget-1.10.0.smith-plugin.zip", "sha256": "` + digestA + `"},
        "2.0.0-beta.1": {"url": "https://example.com/widget-2.0.0-beta.1.smith-plugin.zip", "sha256": "` + digestA + `"}
      }
    },
    "deployer": {
      "description": "Deploys to the moon",
      "roles": ["deploy"], "license": "Apache-2.0",
      "versions": {"0.2.0": {"url": "https://example.com/deployer-0.2.0.smith-plugin.zip", "sha256": "` + digestA + `"}}
    }
  }
}`

func mustIndex(t *testing.T, text string) Index {
	t.Helper()
	idx, err := ParseIndex([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return idx
}

func TestParseIndexFillsNamesAndVersions(t *testing.T) {
	idx := mustIndex(t, sampleIndexJSON)
	w := idx.Plugins["widget"]
	if w.Name != "widget" || w.Versions["1.0.0"].Version != "1.0.0" || w.Description == "" {
		t.Fatalf("got %+v", w)
	}
}

func TestParseIndexRefusesWhatInstallCouldNotTrust(t *testing.T) {
	cases := map[string]struct{ text, want string }{
		"schema":        {`{"schema":2,"plugins":{}}`, "schema 2"},
		"json":          {`{`, "not valid JSON"},
		"name":          {strings.Replace(sampleIndexJSON, `"widget"`, `"../evil"`, 1), "plugin name"},
		"no version":    {`{"schema":1,"plugins":{"a":{"description":"d","license":"MIT","versions":{}}}}`, "no version"},
		"short digest":  {strings.Replace(sampleIndexJSON, digestA, "abc", 1), "64 lowercase hex"},
		"upper digest":  {strings.Replace(sampleIndexJSON, digestA, strings.ToUpper(digestA), 1), "64 lowercase hex"},
		"plain http":    {strings.Replace(sampleIndexJSON, "https://example.com/widget-1.0.0", "http://example.com/widget-1.0.0", 1), "only https"},
		"file url":      {strings.Replace(sampleIndexJSON, "https://example.com/widget-1.0.0", "file:///widget-1.0.0", 1), "not a URL"},
		"bad version":   {strings.Replace(sampleIndexJSON, `"1.0.0"`, `"1/0"`, 1), "version"},
		"no url scheme": {strings.Replace(sampleIndexJSON, "https://example.com/widget-1.0.0", "widget-1.0.0", 1), "URL"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseIndex([]byte(c.text))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want an error with %q, got %v", c.want, err)
			}
		})
	}
}

func TestParseIndexAcceptsHTTPOnThisMachine(t *testing.T) {
	text := strings.Replace(sampleIndexJSON, "https://example.com/widget-1.0.0", "http://127.0.0.1:9/widget-1.0.0", 1)
	if _, err := ParseIndex([]byte(text)); err != nil {
		t.Fatal(err)
	}
}
