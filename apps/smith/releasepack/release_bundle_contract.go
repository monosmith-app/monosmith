package releasepack

// BundleFile and BundleMarkdown are the names of the two bundle assets of a
// GitHub Release: the machine-readable list and the human instructions.
const (
	BundleFile     = "release-bundle.json"
	BundleMarkdown = "INSTALL.md"
	// ChecksumSuffix ends the name of the file that holds a package file's
	// SHA-256: core-0.5.2.tgz.sha256, as the other release assets do.
	ChecksumSuffix = ".sha256"
)

// File is one package file attached to a release.
type File struct {
	// Name is the asset name: core-0.5.2.tgz.
	Name string `json:"name"`
	// SHA256 is the lowercase hex digest of the file.
	SHA256 string `json:"sha256"`
	// Size is the length in bytes.
	Size int64 `json:"size"`
}

// Package is one project's packed files of a release run.
type Package struct {
	Project string `json:"project"`
	// Ecosystem is npm, Python, NuGet, RubyGems or crates.
	Ecosystem string `json:"ecosystem"`
	Version   string `json:"version"`
	// Tag is the release tag whose GitHub Release holds the files.
	Tag   string `json:"tag"`
	Files []File `json:"files"`
}

// Install is how one ecosystem's packages of the run are fetched and
// installed together.
type Install struct {
	Ecosystem string `json:"ecosystem"`
	// Download lists one gh command per release tag that holds files of the
	// ecosystem; run them in one folder.
	Download []string `json:"download"`
	// Command installs every package of the ecosystem from that folder. It
	// has one line per command.
	Command string `json:"command"`
}

// Bundle is the content of release-bundle.json: every package of one release
// run, with the install commands. The same bundle is attached to the release
// of each package in it.
type Bundle struct {
	// Repository is owner/name, for the gh commands.
	Repository string    `json:"repository"`
	Packages   []Package `json:"packages"`
	Install    []Install `json:"install"`
}
