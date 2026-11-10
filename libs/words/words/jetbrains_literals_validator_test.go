package words

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// rawSentence matches a string literal of 12 or more characters that starts with a capital
// letter and holds no escape or template start: wording that belongs in the catalogue.
var rawSentence = regexp.MustCompile(`"([A-Z][^"\\$]{11,})"`)

// parsingContext marks a line whose literal is matched against smith output or is code, not
// text for a person.
var parsingContext = regexp.MustCompile(`Regex\(|startsWith\(|endsWith\(|\.contains\(|== "|!= "|removePrefix\(|removeSuffix\(|@Suppress`)

// jetbrainsLiteralAllowList names the raw literals that stay in the Kotlin code, with the reason.
// An entry is a path below the Kotlin root and the start of the literal.
var jetbrainsLiteralAllowList = []struct{ file, literal, reason string }{
	{"dev/monosmith/jetbrains/composestack/ComposePickDialogHandler.kt", "Tick what the project is made of", "default text inside a multi-line HTML template; migrate with the dialog"},
}

// The JetBrains plugin shows wording from the catalogue only: a new sentence in a handler or a
// use case must be a key, as the Kotlin build cannot be run by this test.
func TestNoRawSentenceInTheJetBrainsKotlinCode(t *testing.T) {
	root := filepath.Join("..", "..", "..", "apps", "jetbrains", "src", "main", "kotlin")
	if _, err := os.Stat(root); err != nil {
		t.Skip("the JetBrains plugin is not in this checkout")
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".kt") {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "dev/monosmith/jetbrains/words/") {
			return nil // generated accessor
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(body), "\n") {
			code := strings.TrimSpace(line)
			if strings.HasPrefix(code, "*") || strings.HasPrefix(code, "//") || strings.HasPrefix(code, "/*") || parsingContext.MatchString(code) {
				continue
			}
			for _, m := range rawSentence.FindAllStringSubmatch(code, -1) {
				if !allowedLiteral(rel, m[1]) {
					t.Errorf("%s:%d: raw wording %q: add a key to libs/words/catalog/en/jetbrains_*.json and call SmithWordsConfig.say", rel, i+1, m[1])
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func allowedLiteral(file, literal string) bool {
	for _, a := range jetbrainsLiteralAllowList {
		if a.file == file && strings.HasPrefix(literal, a.literal) {
			return true
		}
	}
	return false
}
