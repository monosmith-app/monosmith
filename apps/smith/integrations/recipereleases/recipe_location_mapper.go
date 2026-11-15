package recipereleases

import (
	"net/url"
	"regexp"
	"strings"
)

// tagPattern is a recipe's release tag: name@x.y.z.
var tagPattern = regexp.MustCompile(`^(.+)@(\d+\.\d+\.\d+)$`)

// Location is a recipe URL that points at one release of a recipe in a
// GitHub repository.
type Location struct {
	URL         string
	Owner, Repo string
	// Name and Version are those of the tag the URL points at.
	Name, Version string
	// segment is the tag as the URL spells it (team@1.2.0 or team%401.2.0).
	segment string
}

// Parse reads a recipe URL that names a release tag of a GitHub repository:
//
//	https://github.com/<owner>/<repo>/releases/download/<name>@<x.y.z>/<file>
//	https://github.com/<owner>/<repo>/raw/<name>@<x.y.z>/<path>
//	https://raw.githubusercontent.com/<owner>/<repo>/<name>@<x.y.z>/<path>
//
// (the tag may also follow refs/tags/, and '@' may be written %40). Any
// other URL, a branch, or a tag that is not name@x.y.z is not a pinned
// release, and ok is false.
func Parse(ref string) (Location, bool) {
	u, err := url.Parse(ref)
	if err != nil || u.Scheme != "https" {
		return Location{}, false
	}
	segs := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	var rest []string
	switch u.Host {
	case "github.com":
		switch {
		case len(segs) > 5 && segs[2] == "releases" && segs[3] == "download":
			rest = segs[4:]
		case len(segs) > 4 && (segs[2] == "raw" || segs[2] == "blob"):
			rest = segs[3:]
		}
	case "raw.githubusercontent.com":
		if len(segs) > 3 {
			rest = segs[2:]
		}
	}
	if len(rest) == 0 {
		return Location{}, false
	}
	if len(rest) > 3 && rest[0] == "refs" && rest[1] == "tags" {
		rest = rest[2:]
	}
	tag, err := url.PathUnescape(rest[0])
	if err != nil {
		return Location{}, false
	}
	m := tagPattern.FindStringSubmatch(tag)
	if m == nil {
		return Location{}, false
	}
	return Location{URL: ref, Owner: segs[0], Repo: segs[1], Name: m[1], Version: m[2], segment: rest[0]}, true
}

// Pinned returns the URL pointing at another version of the same recipe.
func (l Location) Pinned(version string) string {
	tag := l.Name + "@" + version
	if strings.Contains(l.segment, "%40") {
		tag = l.Name + "%40" + version
	}
	return strings.Replace(l.URL, l.segment, tag, 1)
}
