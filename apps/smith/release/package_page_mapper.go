package release

// packagePage is the page of a published package version on its registry, or
// "" for a registry smith does not know.
func packagePage(kind, name, version string) string {
	switch kind {
	case KindNPM:
		return "https://www.npmjs.com/package/" + name + "/v/" + version
	case KindPyPI:
		return "https://pypi.org/project/" + name + "/" + version
	case KindNuGet:
		return "https://www.nuget.org/packages/" + name + "/" + version
	case KindRubyGems:
		return "https://rubygems.org/gems/" + name + "/versions/" + version
	}
	return ""
}
