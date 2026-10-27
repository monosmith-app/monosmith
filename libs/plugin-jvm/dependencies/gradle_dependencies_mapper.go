package dependencies

import "regexp"

// SectionGradle is where a Gradle build script declares a dependency in the
// string notation ("group:artifact:version").
const SectionGradle = "gradle"

// gradleCoordinate is "group:artifact:version" in quotes, as implementation,
// api, testImplementation and the other configurations write it. A version
// that starts with something other than a digit (a variable, a catalog
// alias, a range) is not matched.
var gradleCoordinate = regexp.MustCompile(`["']([\w.\-]+:[\w.\-]+):(\d[^"':$\s]*)["']`)

// ParseGradle reads the dependencies a build.gradle or build.gradle.kts
// declares in the string notation. Version catalogs (libs.versions.toml),
// the map notation and variables are not read.
func ParseGradle(src []byte) []Coordinate {
	var out []Coordinate
	for _, m := range gradleCoordinate.FindAllSubmatch(src, -1) {
		out = append(out, Coordinate{Section: SectionGradle, Name: string(m[1]), Version: string(m[2])})
	}
	return out
}
