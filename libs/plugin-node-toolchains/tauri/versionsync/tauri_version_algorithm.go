package versionsync

import (
	"regexp"
	"strings"
)

// confVersion finds the top-level "version" of a tauri.conf.json: the span of
// its string value (without the quotes) and the value. ok is false when the
// file has none or it is not a string.
func confVersion(src string) (start, end int, value string, ok bool) {
	depth, i := 0, 0
	expectKey := false
	for i < len(src) {
		c := src[i]
		switch c {
		case '{', '[':
			depth++
			expectKey = c == '{' && depth == 1
			i++
		case '}', ']':
			depth--
			i++
		case ',':
			expectKey = depth == 1
			i++
		case '"':
			s, e := i+1, i+1
			for e < len(src) && src[e] != '"' {
				if src[e] == '\\' {
					e++
				}
				e++
			}
			if e >= len(src) {
				return 0, 0, "", false
			}
			i = e + 1
			if depth != 1 || !expectKey || src[s:e] != "version" {
				expectKey = false
				continue
			}
			rest := strings.TrimLeft(src[i:], " \t\r\n")
			if !strings.HasPrefix(rest, ":") {
				return 0, 0, "", false
			}
			rest = strings.TrimLeft(rest[1:], " \t\r\n")
			if !strings.HasPrefix(rest, `"`) {
				return 0, 0, "", false
			}
			vs := len(src) - len(rest) + 1
			ve := strings.IndexByte(src[vs:], '"')
			if ve < 0 {
				return 0, 0, "", false
			}
			return vs, vs + ve, src[vs : vs+ve], true
		default:
			i++
		}
	}
	return 0, 0, "", false
}

// isVersionPath reports whether a tauri.conf.json version is a path to the
// file that holds the version ("../package.json"), which Tauri 2 reads itself.
func isVersionPath(v string) bool { return strings.HasSuffix(v, ".json") }

// setConfVersion writes version into tauri.conf.json's top-level "version"
// unless it is a path to a package.json; changed is false then, or when the
// file has no version.
func setConfVersion(src, version string) (string, bool) {
	s, e, v, ok := confVersion(src)
	if !ok || isVersionPath(v) || v == version {
		return src, false
	}
	return src[:s] + version + src[e:], true
}

var (
	cargoPackageHeader = regexp.MustCompile(`(?m)^\[package\][ \t]*(#.*)?\r?\n?`)
	cargoVersionLine   = regexp.MustCompile(`(?m)^([ \t]*version[ \t]*=[ \t]*)"([^"]*)"`)
	cargoTableHeader   = regexp.MustCompile(`(?m)^\[`)
)

// cargoVersion is the literal version of Cargo.toml's [package], and ok is
// false when it has none (version.workspace = true).
func cargoVersion(src string) (string, bool) {
	body, _, ok := packageBody(src)
	if !ok {
		return "", false
	}
	m := cargoVersionLine.FindStringSubmatch(body)
	if m == nil {
		return "", false
	}
	return m[2], true
}

// setCargoVersion writes version into [package] of a Cargo.toml; changed is
// false when it has no literal version or already holds this one.
func setCargoVersion(src, version string) (string, bool) {
	body, start, ok := packageBody(src)
	if !ok {
		return src, false
	}
	loc := cargoVersionLine.FindStringSubmatchIndex(body)
	if loc == nil || body[loc[4]:loc[5]] == version {
		return src, false
	}
	updated := body[:loc[4]] + version + body[loc[5]:]
	return src[:start] + updated + src[start+len(body):], true
}

// packageBody is the text of [package] up to the next table, and where it starts.
func packageBody(src string) (body string, start int, ok bool) {
	h := cargoPackageHeader.FindStringIndex(src)
	if h == nil {
		return "", 0, false
	}
	end := len(src)
	if next := cargoTableHeader.FindStringIndex(src[h[1]:]); next != nil {
		end = h[1] + next[0]
	}
	return src[h[1]:end], h[1], true
}
