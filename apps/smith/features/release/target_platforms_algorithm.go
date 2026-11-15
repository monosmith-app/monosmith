package release

import (
	"slices"
	"strings"
)

// platformOf infers the operating system and CPU architecture a release file
// is for from its name: an npm platform package (cli-darwin-x64), a wheel tag
// (macosx_12_0_arm64), a .nupkg runtime id (osx-x64), a gem platform
// (x86_64-linux) or an archive (linux_amd64). The system is darwin, linux or
// windows; the architecture is x64, arm64 or x86, or "" when the name has none.
// ok is false when the name holds no system, so the file is for every platform.
func platformOf(name string) (system, arch string, ok bool) {
	lower := strings.ToLower(name)
	switch {
	case containsAny(lower, "darwin", "macosx", "osx"):
		system = "darwin"
	case strings.Contains(lower, "linux"):
		system = "linux"
	case containsAny(lower, "windows", "win32", "win_", "win-", "mingw"):
		system = "windows"
	case strings.HasSuffix(lower, ".deb"), strings.HasSuffix(lower, ".rpm"):
		system = "linux"
	default:
		return "", "", false
	}
	switch {
	case containsAny(lower, "aarch64", "arm64"):
		arch = "arm64"
	case containsAny(lower, "x86_64", "amd64", "x64"):
		arch = "x64"
	case containsAny(lower, "i686", "i386", "ia32", "x86"):
		arch = "x86"
	}
	return system, arch, true
}

func containsAny(text string, parts ...string) bool {
	return slices.ContainsFunc(parts, func(part string) bool { return strings.Contains(text, part) })
}

// systemOrder and archOrder fix the order platforms are listed in, so the
// same release prints the same lines.
var (
	systemOrder = []string{"darwin", "linux", "windows"}
	archOrder   = []string{"x64", "arm64", "x86"}
)

// platformPairs turns file names into the sorted, unique "system/arch" pairs
// (or "system" when the name has no architecture) the names are for.
func platformPairs(names []string) []string {
	seen := map[string]bool{}
	for _, name := range names {
		if system, arch, ok := platformOf(name); ok {
			pair := system
			if arch != "" {
				pair += "/" + arch
			}
			seen[pair] = true
		}
	}
	pairs := make([]string, 0, len(seen))
	for pair := range seen {
		pairs = append(pairs, pair)
	}
	slices.SortFunc(pairs, func(a, b string) int {
		sa, aa, _ := strings.Cut(a, "/")
		sb, ab, _ := strings.Cut(b, "/")
		if c := slices.Index(systemOrder, sa) - slices.Index(systemOrder, sb); c != 0 {
			return c
		}
		return slices.Index(archOrder, aa) - slices.Index(archOrder, ab)
	})
	return pairs
}

// systemName is how a system is named to a reader: macOS, Linux, Windows.
func systemName(system string) string {
	switch system {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	}
	return system
}

// compactPlatforms lists the systems and then the architectures the pairs
// hold: "macOS, Linux, Windows x x64, arm64".
func compactPlatforms(pairs []string) string {
	var systems, archs []string
	for _, pair := range pairs {
		system, arch, _ := strings.Cut(pair, "/")
		if name := systemName(system); !slices.Contains(systems, name) {
			systems = append(systems, name)
		}
		if arch != "" && !slices.Contains(archs, arch) {
			archs = append(archs, arch)
		}
	}
	text := strings.Join(systems, ", ")
	if len(archs) > 0 {
		slices.SortFunc(archs, func(a, b string) int { return slices.Index(archOrder, a) - slices.Index(archOrder, b) })
		text += " x " + strings.Join(archs, ", ")
	}
	return text
}

// detailPlatforms lists each system with its architectures: "darwin x64 arm64
// · linux x64 arm64". npm names Windows win32, so a caller can ask for that.
func detailPlatforms(pairs []string, windows string) string {
	var groups []string
	for _, system := range systemOrder {
		var archs []string
		found := false
		for _, pair := range pairs {
			s, arch, _ := strings.Cut(pair, "/")
			if s != system {
				continue
			}
			found = true
			if arch != "" {
				archs = append(archs, arch)
			}
		}
		if !found {
			continue
		}
		name := system
		if system == "windows" {
			name = windows
		}
		groups = append(groups, strings.Join(append([]string{name}, archs...), " "))
	}
	return strings.Join(groups, " · ")
}
