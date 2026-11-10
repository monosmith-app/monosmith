package linking

import (
	"strings"
	"testing"
)

var internalNames = map[string]bool{"@auto/env": true, "@auto/retry": true, "@auto/tmpl": true, "@auto/x": true}

func TestPinnedInternalRanges(t *testing.T) {
	manifest := []byte(`{
	"name": "@auto/app",
	"dependencies": {
		"@auto/env": "^0.0.1",
		"@auto/retry": "*",
		"@auto/x": "workspace:*",
		"left-pad": "^1.0.0"
	},
	"devDependencies": {
		"@auto/tmpl": "~1.2.3"
	},
	"optionalDependencies": {
		"@auto/x": "2.0.0"
	},
	"peerDependencies": {
		"@auto/retry": "^1.0.0"
	}
}
`)
	got, err := PinnedInternalRanges(manifest, internalNames)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, p := range got {
		lines = append(lines, p.String())
	}
	want := "@auto/env ^0.0.1 (dependencies)|@auto/tmpl ~1.2.3 (devDependencies)|@auto/x 2.0.0 (optionalDependencies)"
	if strings.Join(lines, "|") != want {
		t.Fatalf("got %q want %q", strings.Join(lines, "|"), want)
	}
}

func TestPinnedInternalRangesSkipsSourceSpecs(t *testing.T) {
	for _, spec := range []string{"*", "workspace:^", "workspace:*", "file:../env", "link:../env", "git+https://h/r.git", "https://h/a.tgz", "github:o/r", "o/r"} {
		manifest := []byte(`{"dependencies": {"@auto/env": "` + spec + `"}}`)
		got, err := PinnedInternalRanges(manifest, internalNames)
		if err != nil || len(got) != 0 {
			t.Errorf("%s: got %v, %v", spec, got, err)
		}
	}
}

func TestPinnedInternalRangesRejectsBrokenText(t *testing.T) {
	if _, err := PinnedInternalRanges([]byte(`{`), internalNames); err == nil {
		t.Error("a truncated file must fail")
	}
}

func TestUseAnySpecKeepsEverythingElse(t *testing.T) {
	for name, text := range map[string]string{
		"two spaces": "{\n  \"name\": \"a\",\n  \"dependencies\": {\n    \"@auto/env\": \"^0.0.1\",\n    \"zed\": \"^1.0.0\"\n  },\n  \"devDependencies\": {\n    \"@auto/tmpl\": \"~1.2.3\"\n  }\n}\n",
		"tabs":       "{\n\t\"name\": \"a\",\n\t\"dependencies\": {\n\t\t\"@auto/env\": \"^0.0.1\",\n\t\t\"zed\": \"^1.0.0\"\n\t},\n\t\"devDependencies\": {\n\t\t\"@auto/tmpl\": \"~1.2.3\"\n\t}\n}\n",
		"crlf":       "{\r\n  \"name\": \"a\",\r\n  \"dependencies\": {\r\n    \"@auto/env\": \"^0.0.1\",\r\n    \"zed\": \"^1.0.0\"\r\n  },\r\n  \"devDependencies\": {\r\n    \"@auto/tmpl\": \"~1.2.3\"\r\n  }\r\n}\r\n",
	} {
		pinned, err := PinnedInternalRanges([]byte(text), internalNames)
		if err != nil || len(pinned) != 2 {
			t.Fatalf("%s: %v %v", name, pinned, err)
		}
		got := string(UseAnySpec([]byte(text), pinned))
		want := strings.NewReplacer(`"^0.0.1"`, `"*"`, `"~1.2.3"`, `"*"`).Replace(text)
		if got != want {
			t.Errorf("%s: got %q want %q", name, got, want)
		}
	}
}
