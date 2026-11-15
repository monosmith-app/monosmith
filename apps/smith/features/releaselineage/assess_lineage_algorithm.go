package releaselineage

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Check names, as doctor reports them.
const (
	CheckResolve  = "release tags resolve under each project's tag prefix"
	CheckRegistry = "release tags are not behind the registry"
)

// Assess checks each subject's tags against its old names and its registry.
func Assess(subjects []Subject) []Finding {
	var out []Finding
	// A prefix another project releases under is never an earlier name of this
	// one: two projects may share a package name in different registries
	// (a Python and a Dart package called opencraw), and each owns its tags.
	owner := map[string]string{}
	for _, s := range subjects {
		owner[s.Prefix] = s.Project
	}
	for _, s := range subjects {
		own, ownTag := newest(s.Tags, s.Prefix)
		alt, altTag := "", ""
		for _, prefix := range s.Alternatives {
			if prefix == s.Prefix || (owner[prefix] != "" && owner[prefix] != s.Project) {
				continue
			}
			if v, tag := newest(s.Tags, prefix); v != "" && (alt == "" || less(alt, v)) {
				alt, altTag = v, tag
			}
		}
		switch {
		case alt != "" && (!s.Explicit || own != "") && (own == "" || less(own, alt)):
			out = append(out, Finding{Project: s.Project, Check: CheckResolve, Blocking: true, Message: fmt.Sprintf(
				"%s's newest release tag is %s, under another name than its prefix %s, so the next release would start from %s.",
				s.Project, altTag, s.Prefix, describe(own, s.Prefix)),
				Remedy: []string{
					fmt.Sprintf("tag the same commit under the current name: git tag %s%s %s^{commit}, then push the tag", s.Prefix, alt, altTag),
					fmt.Sprintf("or keep the old name: set [projects.%s] tag_prefix = %q in smith.toml", s.Project, strings.TrimSuffix(altTag, alt)),
				},
				Baseline: &Baseline{Name: s.Prefix + alt, Target: altTag + "^{commit}"}})
			continue
		case s.Unreachable && NotPublishedYet(s.Detail):
			// A registry answers 401 or 404 for a name nobody has published: a
			// new package, not a network fault, and there is nothing to compare
			// its tags with yet.
			continue
		case s.Unreachable:
			out = append(out, Finding{Project: s.Project, Check: CheckRegistry, Message: fmt.Sprintf(
				"could not ask %s about %s (%s), so its tags were not checked against what is published.", orRegistry(s.Registry), s.Project, s.Detail),
				Remedy: []string{"check the network and your registry login, then run this again"}})
			continue
		case s.Published == "" || !plain(s.Published):
			continue
		case own == "":
			out = append(out, Finding{Project: s.Project, Check: CheckRegistry, Blocking: true, Message: fmt.Sprintf(
				"%s is published at %s according to %s, but has no release tag, so the next release would start from its manifest.",
				s.Project, s.Published, orRegistry(s.Registry)),
				Remedy: []string{
					"git fetch --tags, in case the tag exists on the remote",
					fmt.Sprintf("or tag the commit that published it: git tag %s%s <commit>, then push the tag", s.Prefix, s.Published),
					"or, if you never published it, that package is someone else's with the same name: give yours a scope (@you/name) or mark it \"private\": true",
				}})
		case less(own, s.Published):
			out = append(out, Finding{Project: s.Project, Check: CheckRegistry, Blocking: true, Message: fmt.Sprintf(
				"%s is tagged %s but %s already has %s, so the next release would propose a version the registry refuses.",
				s.Project, ownTag, orRegistry(s.Registry), s.Published),
				Remedy: []string{fmt.Sprintf("tag the commit that published it: git tag %s%s <commit>, then push the tag", s.Prefix, s.Published)}})
		}
	}
	return out
}

// NotPublishedYet reports whether a registry's failure detail is the answer a
// registry gives for a package that was never published: HTTP 401 or 404
// (npm's E401, E404, "401 Unauthorized", "404 Not Found"). Private feeds such
// as Azure Artifacts answer 401, not 404, for a name they do not hold.
func NotPublishedYet(detail string) bool {
	return notPublished.MatchString(detail)
}

var notPublished = regexp.MustCompile(`(?i)\bE?40[14]\b|\b40[14] (not found|unauthorized)\b`)

func describe(v, prefix string) string {
	if v == "" {
		return "scratch"
	}
	return prefix + v
}

func orRegistry(r string) string {
	if r == "" {
		return "the registry"
	}
	return r
}

// newest is the highest plain X.Y.Z version tagged under a prefix.
func newest(tags []string, prefix string) (version, tag string) {
	for _, t := range tags {
		v, ok := strings.CutPrefix(t, prefix)
		if !ok || !plain(v) {
			continue
		}
		if version == "" || less(version, v) {
			version, tag = v, t
		}
	}
	return version, tag
}

// plain reports an X.Y.Z version without pre-release or build parts.
func plain(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err != nil || p == "" {
			return false
		}
	}
	return true
}

func less(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		x, _ := strconv.Atoi(as[i])
		y, _ := strconv.Atoi(bs[i])
		if x != y {
			return x < y
		}
	}
	return false
}
