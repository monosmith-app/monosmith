package release

import (
	"cmp"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"
)

// Kinds of release target. A registry the host does not know is "registry".
const (
	KindNPM           = "npm"
	KindPyPI          = "PyPI"
	KindNuGet         = "NuGet"
	KindRubyGems      = "RubyGems"
	KindRegistry      = "registry"
	KindImage         = "image"
	KindGitHubRelease = "GitHub Release"
	KindMirror        = "mirror"
	KindPub           = "pub.dev"
)

// ReleasedTarget is one place a release put something: a registry, an image
// registry, a GitHub Release, a split repository or pub.dev. It is what the
// summary prints and what the --json report holds under "targets".
type ReleasedTarget struct {
	// Kind is the type of target (the Kind constants).
	Kind string `json:"kind"`
	// What is what went there, in words: the packages, the files, the image.
	What string `json:"what"`
	// Platforms are the "system/arch" pairs the files are for, when the names say.
	Platforms []string `json:"platforms,omitempty"`
	// Where is the registry host, or the repository.
	Where string `json:"where,omitempty"`
	// Count is how many packages or files the target holds.
	Count int `json:"count"`
	// detail is a second line of the text summary (npm: the platforms).
	detail string
	// label names the target in one or two words, for the Published column.
	label string
	// pkg is the main package, and fresh how many of the target's packages this
	// run published (the registry already had the rest): what a deployment records.
	pkg   string
	fresh int
}

// defaultHosts are the registries a package kind goes to unless smith.toml
// names another one, which a report does not record.
var defaultHosts = map[string]string{
	KindNPM: "registry.npmjs.org", KindPyPI: "pypi.org", KindNuGet: "nuget.org", KindRubyGems: "rubygems.org",
}

// registryKinds names the ecosystem of a registry host.
var registryKinds = []struct{ host, kind string }{
	{"npmjs.org", KindNPM}, {"yarnpkg.com", KindNPM}, {"pypi.org", KindPyPI}, {"nuget.org", KindNuGet}, {"rubygems.org", KindRubyGems},
}

// RegistryHost is a registry URL as a reader names it: registry.npmjs.org.
func RegistryHost(url string) string {
	_, rest, found := strings.Cut(url, "://")
	if !found {
		rest = url
	}
	return strings.TrimSuffix(rest, "/")
}

var (
	nugetName   = regexp.MustCompile(`^(.+?)\.\d+\.\d+\.\d+[^/]*\.nupkg$`)
	packageSpec = regexp.MustCompile(`^(.+)@[^@/]+$`)
)

// TargetsOf lists, in a fixed order, every target a package's release reached,
// from what the result already holds. A dry run or a failure before any
// publish gives none.
func TargetsOf(res ReportResult) []ReleasedTarget {
	var targets []ReleasedTarget
	if res.Published && res.Registry != "" {
		host := RegistryHost(res.Registry)
		kind := KindRegistry
		for _, k := range registryKinds {
			if strings.Contains(host, k.host) {
				kind = k.kind
			}
		}
		targets = append(targets, ReleasedTarget{Kind: kind, What: cmp.Or(res.Package, res.Project) + " " + res.Next, Where: host, Count: 1, label: host, fresh: 1, pkg: cmp.Or(res.Package, res.Project)})
	}
	groups := map[string][]string{}
	existing := map[string]int{}
	add := func(name string, already bool) {
		kind := packageKind(name)
		groups[kind] = append(groups[kind], name)
		if already {
			existing[kind]++
		}
	}
	for _, name := range res.Packages {
		add(name, false)
	}
	for _, name := range res.ExistingPackages {
		if name != res.PubExisting {
			add(name, true)
		}
	}
	for _, kind := range []string{KindNPM, KindPyPI, KindNuGet, KindRubyGems} {
		if names := groups[kind]; len(names) > 0 {
			targets = append(targets, packageTarget(kind, names, existing[kind], res.Next))
		}
	}
	for _, image := range res.Images {
		targets = append(targets, ReleasedTarget{Kind: KindImage, What: ImageLine(image), Count: 1, label: KindImage})
	}
	if res.GitHubRelease != "" || len(res.Packed) > 0 {
		targets = append(targets, githubTarget(res))
	}
	if res.Mirror != "" {
		targets = append(targets, ReleasedTarget{Kind: KindMirror, What: res.Mirror, Count: 1, label: res.Mirror})
	}
	switch {
	case res.PubDispatch != "":
		targets = append(targets, ReleasedTarget{Kind: KindPub, What: "publish queued (" + PubWorkflow + "), not finished yet", Count: 1, label: "pub.dev (queued)"})
	case res.PubExisting != "":
		targets = append(targets, ReleasedTarget{Kind: KindPub, What: "already published", Count: 1, label: "pub.dev (already there)"})
	}
	return targets
}

// packageKind says which registry a published package file or name is for.
func packageKind(name string) string {
	switch {
	case strings.HasSuffix(name, ".whl"):
		return KindPyPI
	case strings.HasSuffix(name, ".nupkg"):
		return KindNuGet
	case strings.HasSuffix(name, ".gem"):
		return KindRubyGems
	}
	return KindNPM
}

// packageTarget describes the packages of one registry: npm and NuGet name the
// main packages and count the platform ones; wheels and gems are counted with
// their platforms.
func packageTarget(kind string, names []string, already int, version string) ReleasedTarget {
	t := ReleasedTarget{Kind: kind, Where: defaultHosts[kind], Count: len(names), Platforms: platformPairs(names), label: kind, fresh: len(names) - already}
	switch kind {
	case KindNPM, KindNuGet:
		var main []string
		platform := 0
		for _, name := range names {
			if _, _, ok := platformOf(packageID(kind, name)); ok {
				platform++
			} else {
				main = append(main, packageID(kind, name))
			}
		}
		t.What = joinSome(main, 3)
		if len(main) > 0 {
			t.pkg = main[0]
		}
		if platform > 0 {
			if t.What != "" {
				t.What += " + "
			}
			t.What += fmt.Sprintf("%d platform %s", platform, noun(platform, "package", "packages"))
		}
		if len(t.Platforms) > 0 {
			windows := "windows"
			if kind == KindNPM {
				windows = "win32"
			}
			t.detail = detailPlatforms(t.Platforms, windows)
		}
	default:
		what := "wheel"
		if kind == KindRubyGems {
			what = "gem"
		}
		t.What = fmt.Sprintf("%d %s", len(names), noun(len(names), what, what+"s"))
		name, _, _ := strings.Cut(path.Base(names[0]), "-"+version)
		if kind == KindPyPI {
			name, _, _ = strings.Cut(path.Base(names[0]), "-")
		}
		t.pkg = name
		if len(t.Platforms) > 0 {
			t.What += " (" + compactPlatforms(t.Platforms) + ")"
		}
	}
	if already > 0 {
		t.What += fmt.Sprintf(", %d already published", already)
	}
	return t
}

// packageID is a package's name without its version: name@1.2.0 for npm,
// name.1.2.0.nupkg for NuGet.
func packageID(kind, name string) string {
	if kind == KindNuGet {
		if m := nugetName.FindStringSubmatch(path.Base(name)); m != nil {
			return m[1]
		}
		return strings.TrimSuffix(name, ".nupkg")
	}
	if m := packageSpec.FindStringSubmatch(name); m != nil {
		return m[1]
	}
	return name
}

// githubTarget describes the GitHub Release of a tag and the files it holds,
// counted by kind.
func githubTarget(res ReportResult) ReleasedTarget {
	files := res.Attached
	if len(files) == 0 {
		files = append(files, res.Assets...)
		for _, f := range res.Packed {
			files = append(files, f.Name)
		}
	}
	t := ReleasedTarget{Kind: KindGitHubRelease, What: "release " + res.Tag, Where: res.GitHubRelease, Count: len(files), label: "GitHub Release"}
	if len(res.Packed) > 0 {
		t.label = "attached to GitHub Release"
	}
	if len(files) > 0 {
		t.What += fmt.Sprintf(" with %d %s: %s", len(files), noun(len(files), "file", "files"), summariseFiles(files))
		t.Platforms = platformPairs(files)
	}
	return t
}

func noun(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// joinSome joins names with " + ", and names the first few and a count past few.
func joinSome(names []string, few int) string {
	if len(names) <= few {
		return strings.Join(names, " + ")
	}
	return strings.Join(names[:few], " + ") + fmt.Sprintf(" + %d more", len(names)-few)
}

// TargetLabels names each target of a result in one or two words, for the
// Published column: its registry host, "npm", "2 images", "GitHub Release".
// Images count as one label.
func TargetLabels(res ReportResult) []string {
	var labels []string
	images := 0
	for _, t := range TargetsOf(res) {
		if t.Kind != KindImage {
			labels = append(labels, t.label)
			continue
		}
		if images++; images == 1 {
			labels = append(labels, KindImage)
		}
	}
	if images > 0 {
		at := slices.Index(labels, KindImage)
		labels[at] = fmt.Sprintf("%d %s", images, noun(images, "image", "images"))
	}
	return labels
}
