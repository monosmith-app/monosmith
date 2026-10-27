package dependencies

import (
	"regexp"
	"strings"
)

var (
	openTag      = regexp.MustCompile(`<(PackageReference|PackageVersion)\b([^>]*?)(/?)>`)
	includeAttr  = regexp.MustCompile(`\b(?:Include|Update)\s*=\s*"([^"]*)"`)
	versionAttr  = regexp.MustCompile(`(\bVersion\s*=\s*)"[^"]*"`)
	versionChild = regexp.MustCompile(`(<Version>)[^<]*(</Version>)`)
)

// SetVersion sets the version of one package in an MSBuild file's source,
// where element is PackageReference or PackageVersion, and says whether it
// found the package. The version is an attribute (Version="1.0") or a child
// element (<Version>1.0</Version>); only that text changes, so the file's
// formatting, comments and other metadata stay as written. Package ids are
// compared without regard to case, as NuGet does.
func SetVersion(src, element, name, version string) (string, bool) {
	for _, m := range openTag.FindAllStringSubmatchIndex(src, -1) {
		if src[m[2]:m[3]] != element {
			continue
		}
		attrs := src[m[4]:m[5]]
		id := includeAttr.FindStringSubmatch(attrs)
		if id == nil || !strings.EqualFold(id[1], name) {
			continue
		}
		if versionAttr.MatchString(attrs) {
			changed := versionAttr.ReplaceAllString(attrs, `${1}"`+version+`"`)
			return src[:m[4]] + changed + src[m[5]:], true
		}
		if src[m[6]:m[7]] == "/" {
			continue
		}
		end := strings.Index(src[m[1]:], "</"+element+">")
		if end < 0 {
			continue
		}
		body := src[m[1] : m[1]+end]
		if !versionChild.MatchString(body) {
			continue
		}
		return src[:m[1]] + versionChild.ReplaceAllString(body, "${1}"+version+"${2}") + src[m[1]+end:], true
	}
	return src, false
}
