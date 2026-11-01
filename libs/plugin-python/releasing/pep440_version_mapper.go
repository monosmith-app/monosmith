package releasing

import "strings"

// DistributionVersion is how a distribution's file name spells a semantic
// version, which PEP 440 writes without separators: 2.0.0-beta.1 is 2.0.0b1,
// 2.0.0-alpha.3 is 2.0.0a3, 2.0.0-rc.2 is 2.0.0rc2. A version with no
// prerelease part, or one whose prerelease word PEP 440 has no spelling for, is
// returned as it is.
func DistributionVersion(version string) string {
	core, pre, found := strings.Cut(version, "-")
	if !found {
		return version
	}
	word, number, _ := strings.Cut(pre, ".")
	spelling, ok := map[string]string{"alpha": "a", "a": "a", "beta": "b", "b": "b", "rc": "rc", "c": "rc", "pre": "rc", "preview": "rc"}[word]
	if !ok {
		return version
	}
	if number == "" {
		number = "0"
	}
	for _, r := range number {
		if r < '0' || r > '9' {
			return version
		}
	}
	return core + spelling + number
}

// isDistributionOf reports whether a build output's file name is a source or
// wheel distribution of the version, spelled as given or as PEP 440 writes it:
// name-2.0.0b1.tar.gz and name-2.0.0b1-py3-none-any.whl, but not name-2.0.0b10.
func isDistributionOf(fileName, version string) bool {
	for _, v := range []string{version, DistributionVersion(version)} {
		i := strings.Index(fileName, "-"+v)
		if i < 0 {
			continue
		}
		rest := fileName[i+1+len(v):]
		if rest == "" || rest[0] == '-' || strings.HasPrefix(rest, ".tar.gz") || strings.HasPrefix(rest, ".zip") {
			return true
		}
	}
	return false
}
