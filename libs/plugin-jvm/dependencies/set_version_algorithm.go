package dependencies

import (
	"regexp"
	"strings"
)

var (
	dependencyBlock = regexp.MustCompile(`(?s)<dependency>.*?</dependency>`)
	versionElement  = regexp.MustCompile(`(<version>)\s*[^<]*?\s*(</version>)`)
)

func element(src, name string) string {
	m := regexp.MustCompile(`<` + name + `>\s*([^<]*?)\s*</` + name + `>`).FindStringSubmatch(src)
	if m == nil {
		return ""
	}
	return m[1]
}

// SetPomVersion changes the version of one dependency (groupId:artifactId)
// in a pom.xml source, wherever it gives one (<dependencies> and
// <dependencyManagement> alike, so the two never disagree),
// or, when section names a property ("property:junit.version"), the value
// of that property. Only the version text changes. It says whether it found
// the dependency.
func SetPomVersion(src, section, name, version string) (string, bool) {
	if property, ok := strings.CutPrefix(section, SectionPropertyPrefix); ok {
		tag := regexp.MustCompile(`(<` + regexp.QuoteMeta(property) + `>)[^<]*(</` + regexp.QuoteMeta(property) + `>)`)
		if !tag.MatchString(src) {
			return src, false
		}
		return tag.ReplaceAllString(src, "${1}"+version+"${2}"), true
	}
	changed := false
	out := dependencyBlock.ReplaceAllStringFunc(src, func(block string) string {
		if element(block, "groupId")+":"+element(block, "artifactId") != name || !versionElement.MatchString(block) {
			return block
		}
		changed = true
		return versionElement.ReplaceAllString(block, "${1}"+version+"${2}")
	})
	return out, changed
}

// SetGradleVersion changes every "group:artifact:version" string of one
// dependency in a Gradle build script that currently has the version old.
// It says whether it found one.
func SetGradleVersion(src, name, old, version string) (string, bool) {
	pattern := regexp.MustCompile(`(["'])` + regexp.QuoteMeta(name+":"+old) + `(["'])`)
	if !pattern.MatchString(src) {
		return src, false
	}
	return pattern.ReplaceAllString(src, "${1}"+name+":"+version+"${2}"), true
}
