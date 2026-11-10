package toolinstall

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var (
	goModToolchain = regexp.MustCompile(`(?m)^\s*toolchain\s+go(\d+(?:\.\d+)*)`)
	goModGo        = regexp.MustCompile(`(?m)^\s*go\s+(\d+(?:\.\d+)*)`)
	requiresPython = regexp.MustCompile(`(?m)^\s*requires-python\s*=\s*["']([^"']+)["']`)
	atLeast        = regexp.MustCompile(`(?:>=|~=)\s*(\d+(?:\.\d+)*)`)
	channel        = regexp.MustCompile(`(?m)^\s*channel\s*=\s*["']([^"']+)["']`)
	plainVersion   = regexp.MustCompile(`^v?(\d+(?:\.\d+)*)$`)
)

// FindRequirement reads the version a workspace pins for a tool. It looks for
// each pin of the tool in each folder, in order, and the first file that gives
// a version wins.
//
// Parameters: pins are the tool's pin files; dirs are the absolute folders to
// look in (the workspace root first).
// Returns: the requirement and the base name of the file that gave it; the
// zero requirement and "" when no file pins a version.
func FindRequirement(pins []protocol.ToolVersionPin, dirs []string) (Requirement, string) {
	for _, pin := range pins {
		for _, dir := range dirs {
			data, err := os.ReadFile(filepath.Join(dir, pin.File))
			if err != nil {
				continue
			}
			if req := ParsePin(pin.Format, string(data)); !req.Any() {
				return req, pin.File
			}
		}
	}
	return Requirement{}, ""
}

// ParsePin turns the text of a pin file into a requirement. A format it does
// not know, or text it cannot read ("lts/*", "stable"), gives the zero
// requirement: any version counts.
func ParsePin(format, text string) Requirement {
	switch format {
	case "text":
		first, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
		if m := plainVersion.FindStringSubmatch(strings.TrimSpace(first)); m != nil {
			return Requirement{Prefix: m[1]}
		}
	case "go-mod":
		if m := goModToolchain.FindStringSubmatch(text); m != nil {
			return Requirement{Min: m[1]}
		}
		if m := goModGo.FindStringSubmatch(text); m != nil {
			return Requirement{Min: m[1]}
		}
	case "pyproject":
		if m := requiresPython.FindStringSubmatch(text); m != nil {
			if lo := atLeast.FindStringSubmatch(m[1]); lo != nil {
				return Requirement{Min: lo[1]}
			}
		}
	case "global-json":
		var doc struct {
			SDK struct {
				Version string `json:"version"`
			} `json:"sdk"`
		}
		if json.Unmarshal([]byte(text), &doc) == nil && plainVersion.MatchString(doc.SDK.Version) {
			v := doc.SDK.Version
			parts := strings.Split(v, ".")
			prefix := parts[0]
			if len(parts) > 1 {
				prefix += "." + parts[1]
			}
			return Requirement{Prefix: prefix, Min: v}
		}
	case "rust-toolchain":
		value := strings.TrimSpace(text)
		if m := channel.FindStringSubmatch(text); m != nil {
			value = m[1]
		}
		if m := plainVersion.FindStringSubmatch(value); m != nil {
			return Requirement{Min: m[1]}
		}
	}
	return Requirement{}
}
